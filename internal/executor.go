package internal

import (
	"context"
	"time"

	"github.com/pixel365/pulse/internal/model"
)

type CheckExecutor interface {
	Execute(context.Context, func(context.Context) model.CheckExecutionResult) error
}

var _ CheckExecutor = (*CheckExec)(nil)

type CheckExec struct {
	writer   ResultWriter
	interval time.Duration
	jitter   time.Duration
}

func (c *CheckExec) Execute(
	ctx context.Context,
	fn func(context.Context) model.CheckExecutionResult,
) error {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	Sleep(ctx, c.jitter)

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			Sleep(ctx, c.jitter)
			result := fn(ctx)
			//nolint:staticcheck
			if err := c.writer.Write(ctx, result); err != nil {
				//TODO: log
			}
		}
	}
}

func NewCheckExecutor(
	w ResultWriter,
	i time.Duration,
	j time.Duration,
) *CheckExec {
	return &CheckExec{
		writer:   w,
		interval: i,
		jitter:   j,
	}
}
