package check

import (
	"context"

	"github.com/pixel365/pulse/internal/model"
	"github.com/pixel365/pulse/internal/repository/check"
)

var _ CheckStateService = (*State)(nil)

type State struct {
	repo check.CheckStateRepository
}

func (s *State) GetState(ctx context.Context, service string) (*model.CheckState, error) {
	result, err := s.repo.GetCheckState(ctx, service)

	return result, err
}

func (s *State) UpsertState(ctx context.Context, state model.CheckState) error {
	err := s.repo.UpdateCheckState(ctx, state)

	return err
}

func NewStateService(repo check.CheckStateRepository) *State {
	return &State{repo: repo}
}
