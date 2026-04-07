package check

import "github.com/pixel365/pulse/internal/model"

func transition(
	policy model.CheckPolicy,
	currentState *model.CheckState,
	result model.CheckExecutionResult,
) model.CheckState {
	nextState := model.CheckState{
		CheckID:          policy.CheckID,
		ServiceID:        policy.ServiceID,
		CheckType:        policy.CheckType,
		Status:           model.CheckStateUnknown,
		LastExecutionID:  result.ExecutionID,
		LastStatus:       result.Status,
		LastErrorKind:    result.ErrorKind,
		LastErrorMessage: result.ErrorMessage,
		LastDuration:     result.Duration,
		LastDetails:      result.Details,
		UpdatedAt:        result.FinishedAt,
	}

	if currentState != nil {
		nextState.Status = currentState.Status
		nextState.LastSuccessAt = currentState.LastSuccessAt
		nextState.LastFailureAt = currentState.LastFailureAt
		nextState.ConsecutiveSuccesses = currentState.ConsecutiveSuccesses
		nextState.ConsecutiveFailures = currentState.ConsecutiveFailures
	}

	failureThreshold := minOne(policy.FailureThreshold)

	switch result.Status {
	case model.CheckExecutionSuccess:
		nextState.ConsecutiveSuccesses++
		nextState.ConsecutiveFailures = 0
		nextState.LastSuccessAt = &result.FinishedAt
		nextState.Status = model.CheckStateHealthy
	case model.CheckExecutionFailure:
		nextState.ConsecutiveFailures++
		nextState.ConsecutiveSuccesses = 0
		nextState.LastFailureAt = &result.FinishedAt
		if nextState.ConsecutiveFailures >= failureThreshold {
			nextState.Status = model.CheckStateUnhealthy
		}
	}

	return nextState
}

func minOne(value int) int {
	if value <= 0 {
		return 1
	}

	return value
}
