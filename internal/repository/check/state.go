package check

import (
	"context"

	"github.com/pixel365/pulse/internal/model"
	"github.com/pixel365/pulse/internal/repository"
)

var _ CheckStateRepository = (*StateCheck)(nil)

type StateCheck struct {
	db repository.QueryExecutor
}

func (s *StateCheck) GetCheckState(ctx context.Context, service string) (*model.CheckState, error) {
	return nil, nil
}

func (s *StateCheck) UpdateCheckState(ctx context.Context, state model.CheckState) error {
	return nil
}

func NewStateRepository(db repository.QueryExecutor) *StateCheck {
	return &StateCheck{db}
}
