package execution

import (
	"context"

	"github.com/pixel365/pulse/internal/model"
)

type Service interface {
	ListExecutions(
		context.Context,
		model.CheckExecutionFilter,
	) ([]model.CheckExecutionRecord, error)
	ListExecutionBuckets(
		context.Context,
		model.CheckExecutionAggregateFilter,
	) ([]model.CheckExecutionBucketRecord, error)
	ListExecutionTimeline(
		context.Context,
		model.CheckExecutionTimelineFilter,
	) ([]model.CheckExecutionTimelineRecord, error)
}
