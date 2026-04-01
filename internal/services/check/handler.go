package check

import (
	"context"

	"github.com/pixel365/pulse/internal/model"
)

var _ CheckHandlerService = (*Handler)(nil)

type Handler struct {
	svc CheckStateService
}

func (h *Handler) Handle(
	ctx context.Context,
	policy model.CheckPolicy,
	result model.CheckExecutionResult,
) error {
	return nil
}

func NewHandlerService(svc CheckStateService) *Handler {
	return &Handler{
		svc: svc,
	}
}
