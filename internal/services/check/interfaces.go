package check

import (
	"context"

	"github.com/pixel365/pulse/internal/model"
)

type CheckHandlerService interface {
	Handle(context.Context, model.CheckPolicy, model.CheckExecutionResult) error
}

type CheckStateService interface {
	GetState(context.Context, string) (*model.CheckState, error)
	UpsertState(context.Context, model.CheckState) error
}
