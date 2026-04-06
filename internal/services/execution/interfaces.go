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
}
