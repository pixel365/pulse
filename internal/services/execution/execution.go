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

func (e *Execution) ListExecutionBuckets(
	ctx context.Context,
	filter model.CheckExecutionAggregateFilter,
) ([]model.CheckExecutionBucketRecord, error) {
	return e.repo.ListExecutionBuckets(ctx, filter)
}

func (e *Execution) ListExecutionTimeline(
	ctx context.Context,
	filter model.CheckExecutionTimelineFilter,
) ([]model.CheckExecutionTimelineRecord, error) {
	return e.repo.ListExecutionTimeline(ctx, filter)
}

func NewExecutionService(db repository.QueryExecutor) *Execution {
	return &Execution{
		repo: check.NewExecutionRepository(db),
	}
}
