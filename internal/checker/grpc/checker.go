package grpc

import (
	"context"
	"crypto/rand"
	"time"

	"github.com/pixel365/pulse/internal"
	"github.com/pixel365/pulse/internal/checker"
	c "github.com/pixel365/pulse/internal/config"
	"github.com/pixel365/pulse/internal/model"
)

type Alias = c.TypedCheck[c.GRPCSpec]

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
	return c.executor.Execute(ctx, c.execute)
}

func (c *Checker) execute(ctx context.Context) model.CheckExecutionResult {
	var err error

	result := model.CheckExecutionResult{
		ExecutionID: rand.Text(),
		CheckID:     c.config.ID,
		ServiceID:   c.config.Service,
		CheckType:   c.config.Type,
		Status:      model.Success,
		StartedAt:   time.Now().UTC(),
	}

	attempts := c.config.Retries + 1
	for i := 0; i < attempts; i++ {
		result.AttemptsTotal = i + 1

		if ctx.Err() != nil {
			err = ctx.Err()
			break
		}

		err = nil
		reqErr := c.request(ctx)
		if reqErr != nil {
			err = reqErr
			continue
		}
		break
	}

	result.FinishedAt = time.Now().UTC()
	result.Duration = result.FinishedAt.Sub(result.StartedAt)

	if err != nil {
		result.Status = model.Failure
	} else {
		result.Status = model.Success
	}

	return result
}
