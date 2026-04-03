package check

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/pixel365/pulse/internal/e"
	"github.com/pixel365/pulse/internal/repository"

	"github.com/pixel365/pulse/internal/model"
)

var _ CheckStateRepository = (*StateCheck)(nil)

type StateCheck struct {
	db repository.QueryExecutor
}

func (s *StateCheck) GetCheckState(
	ctx context.Context,
	checkID string,
	serviceID string,
) (*model.CheckState, error) {
	query := `
SELECT
    check_type,
    status,
    last_execution_id,
    last_status,
    last_error_kind,
    last_error_message,
    last_duration,
    last_details,
    last_success_at,
    last_failure_at,
    consecutive_successes,
    consecutive_failures,
    updated_at
FROM pulse.check_states
WHERE check_id = $1 AND service_id = $2
`

	state := model.CheckState{
		CheckID:   checkID,
		ServiceID: serviceID,
	}

	var (
		rawDetails      []byte
		rawErrorKind    *string
		rawErrorMessage *string
		rawDurationUs   int64
	)

	err := s.db.QueryRow(ctx, query, checkID, serviceID).Scan(
		&state.CheckType,
		&state.Status,
		&state.LastExecutionID,
		&state.LastStatus,
		&rawErrorKind,
		&rawErrorMessage,
		&rawDurationUs,
		&rawDetails,
		&state.LastSuccessAt,
		&state.LastFailureAt,
		&state.ConsecutiveSuccesses,
		&state.ConsecutiveFailures,
		&state.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, e.ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	if rawErrorKind != nil {
		state.LastErrorKind = e.ErrorKind(*rawErrorKind)
	}

	if rawErrorMessage != nil {
		state.LastErrorMessage = *rawErrorMessage
	}
	state.LastDuration = time.Duration(rawDurationUs) * time.Microsecond

	if rawDetails != nil {
		if err = json.Unmarshal(rawDetails, &state.LastDetails); err != nil {
			return nil, err
		}
	}

	return &state, nil
}

func (s *StateCheck) UpsertCheckState(
	ctx context.Context,
	state *model.CheckState,
) error {
	if state == nil {
		return errors.New("check state is nil")
	}

	query := `
INSERT INTO pulse.check_states (
    check_id,
    service_id,
    check_type,
    status,
    last_execution_id,
    last_status,
    last_error_kind,
    last_error_message,
    last_duration,
    last_details,
    last_success_at,
    last_failure_at,
    consecutive_successes,
    consecutive_failures,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
)
ON CONFLICT (check_id, service_id) DO UPDATE SET
    check_type = EXCLUDED.check_type,
    status = EXCLUDED.status,
    last_execution_id = EXCLUDED.last_execution_id,
    last_status = EXCLUDED.last_status,
    last_error_kind = EXCLUDED.last_error_kind,
    last_error_message = EXCLUDED.last_error_message,
    last_duration = EXCLUDED.last_duration,
    last_details = EXCLUDED.last_details,
    last_success_at = EXCLUDED.last_success_at,
    last_failure_at = EXCLUDED.last_failure_at,
    consecutive_successes = EXCLUDED.consecutive_successes,
    consecutive_failures = EXCLUDED.consecutive_failures,
    updated_at = EXCLUDED.updated_at
`

	var (
		rawDetails      []byte
		rawErrorKind    *string
		rawErrorMessage *string
	)

	if state.LastDetails != nil {
		data, err := json.Marshal(state.LastDetails)
		if err != nil {
			return err
		}
		rawDetails = data
	}

	if state.LastErrorKind != e.ErrNone {
		rawErrorKind = new(string(state.LastErrorKind))
	}

	if state.LastErrorMessage != "" {
		rawErrorMessage = &state.LastErrorMessage
	}

	_, err := s.db.Exec(ctx, query,
		state.CheckID,
		state.ServiceID,
		state.CheckType,
		state.Status,
		state.LastExecutionID,
		state.LastStatus,
		rawErrorKind,
		rawErrorMessage,
		state.LastDuration.Microseconds(),
		rawDetails,
		state.LastSuccessAt,
		state.LastFailureAt,
		state.ConsecutiveSuccesses,
		state.ConsecutiveFailures,
		state.UpdatedAt,
	)

	return err
}

func (s *StateCheck) ListCheckStatesByService(
	ctx context.Context,
	serviceID string,
) ([]model.CheckState, error) {
	query := `
SELECT 
    check_id,
    check_type,
    status,
    last_execution_id,
    last_status,
    last_error_kind,
    last_error_message,
    last_duration,
    last_details,
    last_success_at,
    last_failure_at,
    consecutive_successes,
    consecutive_failures,
    updated_at
FROM pulse.check_states
WHERE service_id = $1
ORDER BY check_id
`

	var result []model.CheckState
	rows, err := s.db.Query(ctx, query, serviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			rawDetails      []byte
			rawErrorKind    *string
			rawErrorMessage *string
			rawDurationUs   int64
		)
		state := model.CheckState{ServiceID: serviceID}
		err = rows.Scan(
			&state.CheckID,
			&state.CheckType,
			&state.Status,
			&state.LastExecutionID,
			&state.LastStatus,
			&rawErrorKind,
			&rawErrorMessage,
			&rawDurationUs,
			&rawDetails,
			&state.LastSuccessAt,
			&state.LastFailureAt,
			&state.ConsecutiveSuccesses,
			&state.ConsecutiveFailures,
			&state.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if rawErrorKind != nil {
			state.LastErrorKind = e.ErrorKind(*rawErrorKind)
		}

		if rawErrorMessage != nil {
			state.LastErrorMessage = *rawErrorMessage
		}

		if rawDetails != nil {
			if err = json.Unmarshal(rawDetails, &state.LastDetails); err != nil {
				return nil, err
			}
		}

		state.LastDuration = time.Duration(rawDurationUs) * time.Microsecond

		result = append(result, state)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func NewStateRepository(db repository.QueryExecutor) *StateCheck {
	return &StateCheck{db}
}
