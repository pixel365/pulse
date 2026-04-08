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

	var errKind *string
	if result.ErrorKind != e2.ErrNone {
		errKind = new(result.ErrorKind.String())
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
		errKind,
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

func (e *ExecutionCheck) ListExecutionTimeline(
	ctx context.Context,
	filter model.CheckExecutionTimelineFilter,
) ([]model.CheckExecutionTimelineRecord, error) {
	if err := filter.Validate(); err != nil {
		return nil, err
	}

	bucketStepUs, err := executionBucketStepMicros(filter.Bucket)
	if err != nil {
		return nil, err
	}
	staleAfterUs := 2 * filter.Interval.Microseconds()

	query := `
WITH params AS (
      SELECT
          $1::text AS service_id,
          $2::text AS check_id,
          $3::timestamptz AS from_ts,
          $4::timestamptz AS to_ts,
          $5::bigint AS bucket_step_us,
          $6::bigint AS stale_after_us
  ),
  buckets AS (
      SELECT
          gs AS bucket_start,
          LEAST(
              gs + (p.bucket_step_us * interval '1 microsecond'),
              p.to_ts
          ) AS bucket_end
      FROM params p,
      LATERAL generate_series(
          p.from_ts,
          p.to_ts,
          p.bucket_step_us * interval '1 microsecond'
      ) AS gs
      WHERE gs < p.to_ts
  )
  SELECT
      b.bucket_start,
      b.bucket_end,
      last_event.observed_at AS last_observed_at,
      last_event.last_status AS last_execution_status,
      CASE
          WHEN last_event.observed_at IS NULL THEN 'unknown'::pulse.check_state_status
          WHEN b.bucket_end - last_event.observed_at > (
              (SELECT stale_after_us FROM params) * interval '1 microsecond'
          ) THEN 'unknown'::pulse.check_state_status
          ELSE last_event.status
      END AS timeline_state
  FROM buckets b
  LEFT JOIN LATERAL (
      SELECT
          e.observed_at,
          e.last_status,
          e.status
      FROM pulse.check_state_events e
      CROSS JOIN params p
      WHERE e.service_id = p.service_id
        AND e.check_id = p.check_id
        AND e.observed_at < b.bucket_end
      ORDER BY e.observed_at DESC, e.id DESC
      LIMIT 1
  ) last_event ON TRUE
  ORDER BY b.bucket_start
`

	rows, err := e.db.Query(
		ctx,
		query,
		filter.ServiceID,
		filter.CheckID,
		filter.From,
		filter.To,
		bucketStepUs,
		staleAfterUs,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []model.CheckExecutionTimelineRecord

	for rows.Next() {
		var (
			row           model.CheckExecutionTimelineRecord
			rawObservedAt *time.Time
			rawExecStatus *string
		)

		err = rows.Scan(
			&row.BucketStart,
			&row.BucketEnd,
			&rawObservedAt,
			&rawExecStatus,
			&row.State,
		)
		if err != nil {
			return nil, err
		}

		row.LastObservedAt = rawObservedAt
		if rawExecStatus != nil {
			row.LastExecutionStatus = new(model.CheckExecutionStatus(*rawExecStatus))
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

func executionBucketStepMicros(bucket model.CheckExecutionBucket) (int64, error) {
	switch bucket {
	case model.CheckExecutionBucketSecond:
		return time.Second.Microseconds(), nil
	case "", model.CheckExecutionBucketMinute:
		return time.Minute.Microseconds(), nil
	case model.CheckExecutionBucketHour:
		return time.Hour.Microseconds(), nil
	case model.CheckExecutionBucketDay:
		return (24 * time.Hour).Microseconds(), nil
	default:
		return 0, fmt.Errorf("unsupported execution bucket %q", bucket)
	}
}

func NewExecutionRepository(db repository.QueryExecutor) *ExecutionCheck {
	return &ExecutionCheck{db}
}
