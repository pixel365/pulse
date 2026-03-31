package tcp

import (
	"context"

	"github.com/pixel365/pulse/internal"
	"github.com/pixel365/pulse/internal/checker"
	c "github.com/pixel365/pulse/internal/config"
)

type Alias = c.TypedCheck[c.TCPSpec]

var _ checker.Checker = (*Checker)(nil)

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

func (c *Checker) Run(ctx context.Context) error {
	return c.executor.Execute(ctx, c.request)
}
