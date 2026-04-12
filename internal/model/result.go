package model

import (
	"fmt"
	"time"

	"github.com/pixel365/pulse/internal/config"
	"github.com/pixel365/pulse/internal/e"
)

type CheckExecutionStatus string
type CheckStateStatus string
type ServiceStateStatus string
type CheckExecutionBucket string

const (
	CheckExecutionSuccess CheckExecutionStatus = "success"
	CheckExecutionFailure CheckExecutionStatus = "failure"

	CheckStateUnknown   CheckStateStatus = "unknown"
	CheckStateHealthy   CheckStateStatus = "healthy"
	CheckStateUnhealthy CheckStateStatus = "unhealthy"

	ServiceStateUnknown   ServiceStateStatus = "unknown"
	ServiceStateHealthy   ServiceStateStatus = "healthy"
	ServiceStateUnhealthy ServiceStateStatus = "unhealthy"
	ServiceStateDegraded  ServiceStateStatus = "degraded"

	CheckExecutionBucketSecond CheckExecutionBucket = "second"
	CheckExecutionBucketMinute CheckExecutionBucket = "minute"
	CheckExecutionBucketHour   CheckExecutionBucket = "hour"
	CheckExecutionBucketDay    CheckExecutionBucket = "day"
)

type CheckExecutionResult struct {
	StartedAt     time.Time
	FinishedAt    time.Time
	Details       map[string]any
	ExecutionID   string
	ErrorKind     e.ErrorKind
	CheckID       string
	ServiceID     string
	CheckType     config.CheckType
	Status        CheckExecutionStatus
	ErrorMessage  string
	Duration      time.Duration
	AttemptsTotal int
}

type CheckState struct {
	UpdatedAt            time.Time            `json:"updated_at"`
	LastFailureAt        *time.Time           `json:"last_failure_at"`
	LastSuccessAt        *time.Time           `json:"last_success_at"`
	LastDetails          map[string]any       `json:"last_details"`
	LastStatus           CheckExecutionStatus `json:"last_status"`
	LastExecutionID      string               `json:"last_execution_id"`
	CheckID              string               `json:"check_id"`
	LastErrorKind        e.ErrorKind          `json:"last_error_kind"`
	LastErrorMessage     string               `json:"last_error_message"`
	Status               CheckStateStatus     `json:"status"`
	CheckType            config.CheckType     `json:"check_type"`
	ServiceID            string               `json:"service_id"`
	LastDuration         time.Duration        `json:"last_duration"`
	ConsecutiveSuccesses int                  `json:"consecutive_successes"`
	ConsecutiveFailures  int                  `json:"consecutive_failures"`
}

type ServiceState struct {
	ServiceID       string             `json:"service_id"`
	Status          ServiceStateStatus `json:"status"`
	TotalChecks     int                `json:"total_checks"`
	HealthyChecks   int                `json:"healthy_checks"`
	UnhealthyChecks int                `json:"unhealthy_checks"`
	UnknownChecks   int                `json:"unknown_checks"`
}

type CheckPolicy struct {
	CheckID          string
	ServiceID        string
	CheckType        config.CheckType
	FailureThreshold int
}

type CheckExecutionRecord struct {
	CreatedAt time.Time `json:"created_at"`
	CheckExecutionResult
}

type CheckExecutionFilter struct {
	From      *time.Time
	To        *time.Time
	ServiceID string
	CheckID   string
	Limit     int
}

type CheckExecutionAggregateFilter struct {
	Bucket CheckExecutionBucket
	CheckExecutionFilter
}

type CheckExecutionBucketRecord struct {
	BucketStart   time.Time
	Total         int
	SuccessCount  int
	FailureCount  int
	AvgDurationUs int64
}

type CheckExecutionTimelineFilter struct {
	From      time.Time
	To        time.Time
	ServiceID string
	CheckID   string
	Bucket    CheckExecutionBucket
	Interval  time.Duration
}

type CheckExecutionTimelineRecord struct {
	BucketStart         time.Time
	BucketEnd           time.Time
	LastObservedAt      *time.Time
	LastExecutionStatus *CheckExecutionStatus
	State               CheckStateStatus
}

func (f *CheckExecutionFilter) ApplyConditions(fieldFn func(string) string) ([]string, []any) {
	var (
		conditions []string
		args       []any
	)

	if f.ServiceID != "" {
		args = append(args, f.ServiceID)
		conditions = append(conditions, fmt.Sprintf("%s = $%d", fieldFn("service_id"), len(args)))
	}

	if f.CheckID != "" {
		args = append(args, f.CheckID)
		conditions = append(conditions, fmt.Sprintf("%s = $%d", fieldFn("check_id"), len(args)))
	}

	if f.From != nil {
		args = append(args, *f.From)
		conditions = append(conditions, fmt.Sprintf("%s >= $%d", fieldFn("finished_at"), len(args)))
	}

	if f.To != nil {
		args = append(args, *f.To)
		conditions = append(conditions, fmt.Sprintf("%s <= $%d", fieldFn("finished_at"), len(args)))
	}

	return conditions, args
}

func (f *CheckExecutionFilter) Apply(query string, fieldFn func(string) string) (string, []any) {
	conditions, args := f.ApplyConditions(fieldFn)

	if len(conditions) > 0 {
		query += "WHERE " + conditions[0]
		for i := 1; i < len(conditions); i++ {
			query += " AND " + conditions[i]
		}
		query += "\n"
	}

	query += fmt.Sprintf("ORDER BY %s DESC\n", fieldFn("finished_at"))

	if f.Limit > 0 {
		args = append(args, f.Limit)
		query += fmt.Sprintf("LIMIT $%d\n", len(args))
	}

	return query, args
}

func (f *CheckExecutionTimelineFilter) Validate() error {
	if f.ServiceID == "" {
		return fmt.Errorf("service_id is required")
	}

	if f.CheckID == "" {
		return fmt.Errorf("check_id is required")
	}

	if f.From.IsZero() {
		return fmt.Errorf("from is required")
	}

	if f.To.IsZero() {
		return fmt.Errorf("to is required")
	}

	if !f.To.After(f.From) {
		return fmt.Errorf("to must be after from")
	}

	if f.Interval <= 0 {
		return fmt.Errorf("interval must be greater than zero")
	}

	switch f.Bucket {
	case "", CheckExecutionBucketSecond, CheckExecutionBucketMinute,
		CheckExecutionBucketHour, CheckExecutionBucketDay:
	default:
		return fmt.Errorf("unsupported bucket %q", f.Bucket)
	}

	return nil
}
