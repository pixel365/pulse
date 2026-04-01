package check

import (
	"context"

	"github.com/pixel365/pulse/internal/model"
)

type CheckStateRepository interface {
	GetCheckState(context.Context, string) (*model.CheckState, error)
	UpdateCheckState(context.Context, model.CheckState) error
}
