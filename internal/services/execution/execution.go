package execution

import (
	"context"

	"github.com/pixel365/pulse/internal/model"
	"github.com/pixel365/pulse/internal/repository"
	"github.com/pixel365/pulse/internal/repository/check"
)

const (
	defaultExecutionLimit = 100
	maxExecutionLimit     = 1000
)

var _ Service = (*Execution)(nil)

type Execution struct {
	repo check.CheckExecutionRepository
}

func (e *Execution) ListExecutions(
	ctx context.Context,
	filter model.CheckExecutionFilter,
) ([]model.CheckExecutionRecord, error) {
	if filter.Limit <= 0 {
		filter.Limit = defaultExecutionLimit
	}

	if filter.Limit > maxExecutionLimit {
		filter.Limit = maxExecutionLimit
	}

	return e.repo.ListExecutions(ctx, filter)
}

func NewExecutionService(db repository.QueryExecutor) *Execution {
	return &Execution{
		repo: check.NewExecutionRepository(db),
	}
}
