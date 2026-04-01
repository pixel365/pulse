package model

import (
	"time"

	"github.com/pixel365/pulse/internal/config"
	"github.com/pixel365/pulse/internal/e"
)

type CheckExecutionStatus string

const (
	Success CheckExecutionStatus = "success"
	Failure CheckExecutionStatus = "failure"
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
