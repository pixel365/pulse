package http

import (
	"context"
	"time"

	"github.com/pixel365/pulse/internal"
	"github.com/pixel365/pulse/internal/checker"
	c "github.com/pixel365/pulse/internal/config"
	"github.com/pixel365/pulse/internal/model"
)

type Alias = c.TypedCheck[c.HttpSpec]

var _ checker.Checker = (*Checker)(nil)

type Checker struct {
	writer model.ResultWriter
	config Alias
}

func NewChecker(cfg Alias) *Checker {
	return &Checker{
		writer: model.FakeWriter{},
		config: cfg,
	}
}

func (c *Checker) Run(ctx context.Context) error {
	ticker := time.NewTicker(c.config.Interval)
	defer ticker.Stop()

	internal.Sleep(ctx, c.config.Jitter)

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			internal.Sleep(ctx, c.config.Jitter)
			result := c.execute(ctx)
			_ = c.writer.Write(ctx, result)
			//TODO: log
		}
	}
}
