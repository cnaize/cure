package source

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"iter"
	"os"
	"path/filepath"

	"github.com/mholt/archives"
	"github.com/rs/zerolog"

	"github.com/cnaize/cure/logger"
	"github.com/cnaize/cure/source/adapter"
)

var _ Source = (*Local)(nil)

type Local struct {
	path    string
	adapter adapter.Adapter
	logger  *zerolog.Logger
}

func NewLocal(path string) *Local {
	logger := logger.DefaultLogger.
		With().
		Str("module", "source").
		Str("type", "local").
		Logger()

	return &Local{
		path:    path,
		adapter: adapter.NewYara(),
		logger:  &logger,
	}
}

func (s *Local) WithAdapter(adapter adapter.Adapter) *Local {
	s.adapter = adapter

	return s
}

func (s *Local) WithLogger(logger *zerolog.Logger) *Local {
	s.logger = logger

	return s
}

func (s *Local) Files(ctx context.Context) (iter.Seq[io.Reader], error) {
	stat, err := os.Stat(s.path)
	if err != nil {
		return nil, fmt.Errorf("%s: stat: %w", s.path, err)
	}

	return func(yield func(io.Reader) bool) {
		if stat.IsDir() {
			if err := s.dirFiles(yield); err != nil {
				s.logger.Err(err).Str("path", s.path).Msg("failed to walk dir")
			}

			return
		}

		file, err := os.Open(s.path)
		if err != nil {
			s.logger.Err(err).Str("path", s.path).Msg("failed to open path")
			return
		}
		defer file.Close()

		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		if format, stream, err := archives.Identify(ctx, s.path, file); err == nil {
			if extractor, ok := format.(archives.Extractor); ok {
				if err := extractor.Extract(ctx, stream, func(ctx context.Context, info archives.FileInfo) error {
					if info.IsDir() {
						return nil
					}

					if !s.adapter.Check(info.NameInArchive) {
						return nil
					}

					file, err := info.Open()
					if err != nil {
						return fmt.Errorf("%s: open: %w", info.NameInArchive, err)
					}
					defer file.Close()

					if !yield(s.adapter.Adapt(file)) {
						cancel()
						return context.Canceled
					}

					return nil
				}); err != nil {
					s.logger.Err(err).Str("path", s.path).Msg("failed to extract archive")
					return
				}
			}
		} else {
			yield(stream)
		}
	}, nil
}

func (s *Local) dirFiles(yield func(io.Reader) bool) error {
	return filepath.WalkDir(s.path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			s.logger.Err(err).Str("path", s.path).Msg("failed to walk dir entry")
			return nil
		}

		if d.IsDir() {
			return nil
		}

		if !s.adapter.Check(path) {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			s.logger.Err(err).Str("path", s.path).Msg("failed to walk dir file")
			return nil
		}
		defer file.Close()

		if !yield(s.adapter.Adapt(file)) {
			return filepath.SkipAll
		}

		return nil
	})
}
