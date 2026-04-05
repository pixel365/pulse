package check

import (
	"context"
	"encoding/json"
	"fmt"
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

func (e *ExecutionCheck) ListExecutionBuckets(
	ctx context.Context,
	filter model.CheckExecutionAggregateFilter,
) ([]model.CheckExecutionBucketRecord, error) {
	bucketExpr, err := executionBucketExpr(filter.Bucket)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf(`
SELECT
    %s AS bucket_start,
    COUNT(1) AS total,
    COUNT(1) FILTER (WHERE status = 'success') AS success_count,
    COUNT(1) FILTER (WHERE status = 'failure') AS failure_count,
    AVG(duration)::bigint AS avg_duration_us
FROM pulse.check_executions
`, bucketExpr)

	conditions, args := filter.ApplyConditions(mustField)
	if len(conditions) > 0 {
		query += "WHERE " + conditions[0]
		for i := 1; i < len(conditions); i++ {
			query += " AND " + conditions[i]
		}
		query += "\n"
	}

	query += "GROUP BY bucket_start\n"
	query += "ORDER BY bucket_start\n"

	rows, err := e.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []model.CheckExecutionBucketRecord

	for rows.Next() {
		var row model.CheckExecutionBucketRecord

		err = rows.Scan(
			&row.BucketStart,
			&row.Total,
			&row.SuccessCount,
			&row.FailureCount,
			&row.AvgDurationUs,
		)
		if err != nil {
			return nil, err
		}

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

func executionBucketExpr(bucket model.CheckExecutionBucket) (string, error) {
	switch bucket {
	case model.CheckExecutionBucketSecond:
		return "date_trunc('second', finished_at)", nil
	case "", model.CheckExecutionBucketMinute:
		return "date_trunc('minute', finished_at)", nil
	case model.CheckExecutionBucketHour:
		return "date_trunc('hour', finished_at)", nil
	case model.CheckExecutionBucketDay:
		return "date_trunc('day', finished_at)", nil
	default:
		return "", fmt.Errorf("unsupported execution bucket %q", bucket)
	}
}

func NewExecutionRepository(db repository.QueryExecutor) *ExecutionCheck {
	return &ExecutionCheck{db}
}
