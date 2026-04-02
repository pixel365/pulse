package check

import (
	"context"
	"encoding/json"

	e2 "github.com/pixel365/pulse/internal/e"
	"github.com/pixel365/pulse/internal/model"
	"github.com/pixel365/pulse/internal/repository"
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
		result.Duration,
		result.AttemptsTotal,
		&errKind,
		result.ErrorMessage,
		details,
	)

	return err
}

func NewExecutionRepository(db repository.QueryExecutor) *ExecutionCheck {
	return &ExecutionCheck{db}
}
