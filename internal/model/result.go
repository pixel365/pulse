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
	UpdatedAt            time.Time
	LastFailureAt        *time.Time
	LastSuccessAt        *time.Time
	LastDetails          map[string]any
	LastStatus           CheckExecutionStatus
	LastExecutionID      string
	CheckID              string
	LastErrorKind        e.ErrorKind
	LastErrorMessage     string
	Status               CheckStateStatus
	CheckType            config.CheckType
	ServiceID            string
	LastDuration         time.Duration
	ConsecutiveSuccesses int
	ConsecutiveFailures  int
}

type CheckPolicy struct {
	CheckID          string
	ServiceID        string
	CheckType        config.CheckType
	FailureThreshold int
	SuccessThreshold int
}

type CheckExecutionRecord struct {
	CreatedAt time.Time
	ID        string
	CheckExecutionResult
}

type CheckExecutionFilter struct {
	From      *time.Time
	To        *time.Time
	ServiceID string
	CheckID   string
	Limit     int
}

func (f *CheckExecutionFilter) Apply(query string, fieldFn func(string) string) (string, []any) {
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
