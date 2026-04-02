package check

import (
	"context"

	"github.com/pixel365/pulse/internal/model"
	"github.com/pixel365/pulse/internal/repository/check"
)

var _ CheckHandlerService = (*Handler)(nil)

type Handler struct {
	repo check.CheckExecutionRepository
	svc  CheckStateService
}

func (h *Handler) Handle(
	ctx context.Context,
	policy model.CheckPolicy,
	result model.CheckExecutionResult,
) error {
	return h.repo.Add(ctx, result)
}

func NewHandlerService(svc CheckStateService, repo check.CheckExecutionRepository) *Handler {
	return &Handler{
		svc:  svc,
		repo: repo,
	}
}
