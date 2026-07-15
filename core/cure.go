package core

import (
	"context"
	"fmt"
	"io"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog"
	"github.com/sansecio/yargo/scanner"

	"github.com/cnaize/cure/logger"
	"github.com/cnaize/cure/source"
	"github.com/cnaize/cure/source/adapter"
)

type Cure struct {
	sources []source.Source
	options *Options
	logger  *zerolog.Logger

	rules atomic.Pointer[scanner.Rules]
}

func NewCure() *Cure {
	logger := logger.DefaultLogger.
		With().
		Str("module", "core").
		Logger()

	return &Cure{
		// default sources
		sources: []source.Source{
			source.NewRemote("https://github.com/coreruleset/coreruleset/releases/latest/download/coreruleset-4.28.0-minimal.zip").
				WithAdapter(adapter.NewCrs()),
		},
		options: &Options{},
		logger:  &logger,
	}
}

func (c *Cure) WithSources(sources ...source.Source) *Cure {
	c.sources = sources

	return c
}

func (c *Cure) WithOptions(options *Options) *Cure {
	c.options = options

	return c
}

func (c *Cure) WithLogger(logger *zerolog.Logger) *Cure {
	c.logger = logger

	return c
}

func (c *Cure) Scan(data []byte, flags scanner.ScanFlags, timeout time.Duration, cb scanner.ScanCallback) error {
	if len(data) < 1 {
		return nil
	}

	rules := c.rules.Load()
	if rules == nil {
		return nil
	}

	c.logger.Debug().Bytes("data", data).Msg("scanning")

	return rules.ScanMem(data, flags, timeout, cb)
}

func (c *Cure) Run(ctx context.Context) error {
	c.options.SetDefaults()

	go func() {
		ticker := time.NewTicker(c.options.UpdateInterval)
		for {
			if err := c.Update(ctx); err != nil {
				c.logger.Err(err).Msg("failed to update cure")
			}

			select {
			case <-ticker.C:
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

func (c *Cure) Update(ctx context.Context) error {
	rules, err := Compile(func(yield func(io.Reader) bool) {
		for _, source := range c.sources {
			files, err := source.Files(ctx)
			if err != nil {
				c.logger.Err(err).Msg("failed to get source files")
				continue
			}

			for file := range files {
				if !yield(file) {
					return
				}
			}
		}
	})
	if err != nil {
		return fmt.Errorf("compile: %w", err)
	}

	c.rules.Store(rules)

	c.logger.Info().Int("rules", rules.NumRules()).Msg("rules updated")

	return nil
}
