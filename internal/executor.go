package internal

import (
	"context"
	"crypto/rand"
	"time"

	"github.com/pixel365/pulse/internal/config"
	"github.com/pixel365/pulse/internal/e"
	"github.com/pixel365/pulse/internal/model"
	"github.com/pixel365/pulse/internal/services/check"
)

type CheckExecutor interface {
	Execute(context.Context, func(context.Context) error) error
}

var _ CheckExecutor = (*CheckExec)(nil)

type CheckExec struct {
	handler check.CheckHandlerService
	cfg     config.CheckFields
}

func (c *CheckExec) Execute(
	ctx context.Context,
	request func(context.Context) error,
) error {
	ticker := time.NewTicker(c.cfg.Interval)
	defer ticker.Stop()

	Sleep(ctx, c.cfg.Jitter)

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			Sleep(ctx, c.cfg.Jitter)
			result := c.execute(ctx, request)

			policy := model.CheckPolicy{
				CheckID:          result.CheckID,
				ServiceID:        result.ServiceID,
				CheckType:        result.CheckType,
				FailureThreshold: c.cfg.FailureThreshold,
				SuccessThreshold: c.cfg.SuccessThreshold,
			}

			//nolint:staticcheck
			if err := c.handler.Handle(ctx, policy, result); err != nil {
				//TODO: log
			}
		}
	}
}

func (c *CheckExec) execute(
	ctx context.Context,
	request func(context.Context) error,
) model.CheckExecutionResult {
	var err error

	result := model.CheckExecutionResult{
		ExecutionID:  rand.Text(),
		CheckID:      c.cfg.ID,
		ServiceID:    c.cfg.Service,
		CheckType:    c.cfg.Type,
		Status:       model.CheckExecutionSuccess,
		StartedAt:    time.Now().UTC(),
		ErrorKind:    e.ErrNone,
		ErrorMessage: "",
	}

	attempts := c.cfg.Retries + 1
	for i := 0; i < attempts; i++ {
		result.AttemptsTotal = i + 1

		if ctx.Err() != nil {
			err = ctx.Err()
			break
		}

		err = nil
		reqErr := request(ctx)
		if reqErr != nil {
			err = reqErr
			continue
		}
		break
	}

	result.FinishedAt = time.Now().UTC()
	result.Duration = result.FinishedAt.Sub(result.StartedAt)

	if err != nil {
		result.Status = model.CheckExecutionFailure
		result.ErrorKind, result.ErrorMessage = e.ResolveError(err)
	}

	return result
}

func NewCheckExecutor(
	handler check.CheckHandlerService,
	cfg config.CheckFields,
) *CheckExec {
	return &CheckExec{
		handler: handler,
		cfg:     cfg,
	}
}
