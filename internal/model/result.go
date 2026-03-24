package model

import (
	"context"
	"time"

	"github.com/pixel365/pulse/internal/config"
)

type CheckExecutionStatus string

const (
	Success CheckExecutionStatus = "success"
	Failure CheckExecutionStatus = "failure"
)

type ResultWriter interface {
	Write(context.Context, CheckExecutionResult) error
}

type CheckExecutionResult struct {
	ExecutionID   string
	StartedAt     time.Time
	FinishedAt    time.Time
	ErrorKind     ErrorKind
	Details       map[string]any
	CheckID       string
	ServiceID     string
	CheckType     config.CheckType
	Status        CheckExecutionStatus
	Duration      time.Duration
	AttemptsTotal int
}
