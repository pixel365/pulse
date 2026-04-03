package check

import (
	"context"

	"github.com/pixel365/pulse/internal/model"
)

type CheckStateRepository interface {
	GetCheckState(context.Context, string, string) (*model.CheckState, error)
	UpsertCheckState(context.Context, *model.CheckState) error
	ListCheckStatesByService(context.Context, string) ([]model.CheckState, error)
}

type CheckExecutionRepository interface {
	Add(context.Context, model.CheckExecutionResult) error
}
