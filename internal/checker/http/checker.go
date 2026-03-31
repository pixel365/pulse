package http

import (
	"context"

	"github.com/pixel365/pulse/internal"
	c "github.com/pixel365/pulse/internal/config"
)

type Alias = c.TypedCheck[c.HttpSpec]

var _ internal.Checker = (*Checker)(nil)

type Checker struct {
	executor internal.CheckExecutor
	config   Alias
}

func NewChecker(cfg Alias, executor internal.CheckExecutor) *Checker {
	return &Checker{
		executor: executor,
		config:   cfg,
	}
}

func (c *Checker) Check(ctx context.Context) error {
	return c.executor.Execute(ctx, c.request)
}
