package check

import (
	"context"
	"encoding/json"
	"time"

	"github.com/pixel365/pulse/internal/repository"

	e2 "github.com/pixel365/pulse/internal/e"
	"github.com/pixel365/pulse/internal/model"
)

var _ CheckExecutionRepository = (*ExecutionCheck)(nil)

type ExecutionCheck struct {
	db repository.QueryExecutor
}

func (e *ExecutionCheck) Add(ctx context.Context, result model.CheckExecutionResult) error {
	query := `
INSERT INTO pulse.check_executions (
    execution_id,
    check_id,
    service_id,
    status,
    check_type,
    started_at,
    finished_at,
    duration,
    attempts_total,
    error_kind,
    error_message,
    details
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
)
`

	var details []byte
	if result.Details != nil {
		data, err := json.Marshal(result.Details)
		if err != nil {
			return err
		}
		details = data
	}

	var errKind string
	if result.ErrorKind != e2.ErrNone {
		errKind = string(result.ErrorKind)
	}

	_, err := e.db.Exec(ctx, query,
		result.ExecutionID,
		result.CheckID,
		result.ServiceID,
		result.Status,
		result.CheckType,
		result.StartedAt,
		result.FinishedAt,
		result.Duration.Microseconds(),
		result.AttemptsTotal,
		&errKind,
		result.ErrorMessage,
		details,
	)

	return err
}

func (e *ExecutionCheck) ListExecutions(
	ctx context.Context,
	filter model.CheckExecutionFilter,
) ([]model.CheckExecutionRecord, error) {
	query := `
SELECT 
    execution_id,
    check_id,
    service_id,
    status,
    check_type,
    started_at,
    finished_at,
    duration,
    attempts_total,
    error_kind,
    error_message,
    details,
    created_at
FROM pulse.check_executions
`

	query, args := filter.Apply(query, mustField)

	rows, err := e.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []model.CheckExecutionRecord

	for rows.Next() {
		var (
			row             model.CheckExecutionRecord
			details         []byte
			rawErrorKind    *string
			rawErrorMessage *string
			rawDurationUs   int64
		)

		err = rows.Scan(
			&row.ExecutionID,
			&row.CheckID,
			&row.ServiceID,
			&row.Status,
			&row.CheckType,
			&row.StartedAt,
			&row.FinishedAt,
			&rawDurationUs,
			&row.AttemptsTotal,
			&rawErrorKind,
			&rawErrorMessage,
			&details,
			&row.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if rawErrorKind != nil {
			row.ErrorKind = e2.ErrorKind(*rawErrorKind)
		}

		if rawErrorMessage != nil {
			row.ErrorMessage = *rawErrorMessage
		}

		if details != nil {
			err = json.Unmarshal(details, &row.Details)
			if err != nil {
				return nil, err
			}
		}

		row.Duration = time.Duration(rawDurationUs) * time.Microsecond

		result = append(result, row)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func mustField(name string) string {
	switch name {
	case "service_id", "check_id", "finished_at":
		return name
	}

	panic("unknown field " + name + " in check execution filter")
}

func NewExecutionRepository(db repository.QueryExecutor) *ExecutionCheck {
	return &ExecutionCheck{db}
}
