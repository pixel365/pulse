package state

import (
	"context"

	"github.com/pixel365/pulse/internal/model"
)

type StateService interface {
	GetStatesByService(context.Context, string) ([]model.CheckState, error)
}
