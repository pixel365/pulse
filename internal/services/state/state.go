package state

import (
	"context"

	"github.com/pixel365/pulse/internal/model"
	"github.com/pixel365/pulse/internal/repository"
	"github.com/pixel365/pulse/internal/repository/check"
)

var _ StateService = (*State)(nil)

type State struct {
	repo check.CheckStateRepository
}

func (s *State) GetStatesByService(
	ctx context.Context,
	serviceID string,
) ([]model.CheckState, error) {
	return s.repo.ListCheckStatesByService(ctx, serviceID)
}

func NewStateService(db repository.QueryExecutor) *State {
	return &State{
		repo: check.NewStateRepository(db),
	}
}
