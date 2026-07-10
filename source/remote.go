package source

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"iter"
	"path/filepath"
	"slices"
	"strings"

	"github.com/mholt/archives"
	"github.com/rs/zerolog"
	"resty.dev/v3"

	"github.com/cnaize/cure/logger"
)

var _ Source = (*Remote)(nil)

type Remote struct {
	url    string
	logger *zerolog.Logger
}

func NewRemote(url string) *Remote {
	logger := logger.DefaultLogger.
		With().
		Str("module", "source").
		Str("type", "remote").
		Logger()

	return &Remote{
		url:    url,
		logger: &logger,
	}
}

func (s *Remote) WithLogger(logger *zerolog.Logger) *Remote {
	s.logger = logger

	return s
}

func (s *Remote) Files(ctx context.Context) (iter.Seq[io.Reader], error) {
	return func(yield func(io.Reader) bool) {
		resp, err := resty.New().
			SetRetryCount(3).
			AddRetryConditions(func(r *resty.Response, err error) bool {
				return err != nil
			}).
			R().
			WithContext(ctx).
			Get(s.url)
		if err != nil {
			s.logger.Err(err).Str("url", s.url).Msg("failed to get files")
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			s.logger.Err(err).Str("url", s.url).Msg("failed to read body")
			return
		}

		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		if format, stream, err := archives.Identify(ctx, s.url, bytes.NewReader(body)); err == nil {
			if extractor, ok := format.(archives.Extractor); ok {
				if err := extractor.Extract(ctx, stream, func(ctx context.Context, info archives.FileInfo) error {
					if info.IsDir() {
						return nil
					}

					if !slices.Contains(YaraFileExtensions, strings.ToLower(filepath.Ext(info.NameInArchive))) {
						return nil
					}

					file, err := info.Open()
					if err != nil {
						return fmt.Errorf("%s: open: %w", info.NameInArchive, err)
					}
					defer file.Close()

					if !yield(file) {
						cancel()
						return context.Canceled
					}

					return nil
				}); err != nil {
					s.logger.Err(err).Str("url", s.url).Msg("failed to extract archive")
					return
				}
			}
		} else {
			yield(stream)
		}
	}, nil
}
