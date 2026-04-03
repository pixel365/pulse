package check

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/pixel365/pulse/internal/e"

	"github.com/pixel365/pulse/internal/repository/check"

	"github.com/pixel365/pulse/internal/repository"

	"github.com/pixel365/pulse/internal/model"
)

var _ CheckHandlerService = (*Handler)(nil)

type Handler struct {
	txManager repository.TxManager
}

func (h *Handler) Handle(
	ctx context.Context,
	policy model.CheckPolicy,
	result model.CheckExecutionResult,
) error {
	err := repository.Tx(ctx, h.txManager, pgx.ReadCommitted,
		func(tx pgx.Tx) error {
			repo := check.NewExecutionRepository(tx)
			return repo.Add(ctx, result)
		},
		func(tx pgx.Tx) error {
			repo := check.NewStateRepository(tx)
			currentState, err := repo.GetCheckState(ctx, policy.CheckID, policy.ServiceID)
			if err != nil && !errors.Is(err, e.ErrNotFound) {
				return err
			}

			return repo.UpsertCheckState(ctx, new(transition(policy, currentState, result)))
		},
	)

	return err
}

func NewHandlerService(txManager repository.TxManager) *Handler {
	return &Handler{txManager}
}
