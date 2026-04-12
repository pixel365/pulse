package state

import (
	"context"
	"slices"

	"github.com/pixel365/pulse/internal/config"
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

func (s *State) ListServiceStates(
	ctx context.Context,
	cfg *config.Config,
) ([]model.ServiceState, error) {
	serviceIDs := make([]string, 0, len(cfg.Services))
	for serviceID := range cfg.Services {
		serviceIDs = append(serviceIDs, serviceID)
	}
	slices.Sort(serviceIDs)

	result := make([]model.ServiceState, 0, len(serviceIDs))
	for _, serviceID := range serviceIDs {
		states, err := s.repo.ListCheckStatesByService(ctx, serviceID)
		if err != nil {
			return nil, err
		}

		result = append(
			result,
			aggregateServiceState(serviceID, checksByService(cfg, serviceID), states),
		)
	}

	return result, nil
}

func aggregateServiceState(
	serviceID string,
	checks map[string]config.CheckFields,
	states []model.CheckState,
) model.ServiceState {
	result := model.ServiceState{
		ServiceID: serviceID,
		Status:    model.ServiceStateUnknown,
	}

	if len(checks) == 0 {
		return result
	}

	statesByCheckID := make(map[string]model.CheckState, len(states))
	for i := range states {
		statesByCheckID[states[i].CheckID] = states[i]
	}

	hasCriticalUnhealthy := false
	hasNonCriticalUnhealthy := false

	for checkID := range checks {
		result.TotalChecks++
		state, ok := statesByCheckID[checkID]
		if !ok {
			result.UnknownChecks++
			continue
		}

		switch state.Status {
		case model.CheckStateHealthy:
			result.HealthyChecks++
		case model.CheckStateUnhealthy:
			result.UnhealthyChecks++
			if checks[checkID].StatusImpact == config.StatusImpactCritical {
				hasCriticalUnhealthy = true
			} else {
				hasNonCriticalUnhealthy = true
			}
		case model.CheckStateUnknown:
			result.UnknownChecks++
		default:
			result.UnknownChecks++
		}
	}

	switch {
	case hasCriticalUnhealthy:
		result.Status = model.ServiceStateUnhealthy
	case hasNonCriticalUnhealthy:
		result.Status = model.ServiceStateDegraded
	case result.TotalChecks == result.HealthyChecks:
		result.Status = model.ServiceStateHealthy
	default:
		result.Status = model.ServiceStateUnknown
	}

	return result
}

func checksByService(cfg *config.Config, serviceID string) map[string]config.CheckFields {
	result := map[string]config.CheckFields{}

	appendChecksByService(result, serviceID, cfg.HttpChecks)
	appendChecksByService(result, serviceID, cfg.TCPChecks)
	appendChecksByService(result, serviceID, cfg.GRPCChecks)
	appendChecksByService(result, serviceID, cfg.DNSChecks)
	appendChecksByService(result, serviceID, cfg.TLSChecks)

	return result
}

func appendChecksByService[T any](
	dst map[string]config.CheckFields,
	serviceID string,
	checks map[string]config.TypedCheck[T],
) {
	for _, typedCheck := range checks {
		if !typedCheck.Enabled || typedCheck.Service != serviceID {
			continue
		}

		dst[typedCheck.ID] = typedCheck.CheckFields
	}
}

func NewStateService(db repository.QueryExecutor) *State {
	return &State{
		repo: check.NewStateRepository(db),
	}
}
