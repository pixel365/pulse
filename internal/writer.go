package internal

import (
	"context"

	"github.com/pixel365/pulse/internal/model"
)

var _ ResultWriter = (*FakeWriter)(nil)

type ResultWriter interface {
	Write(context.Context, model.CheckExecutionResult) error
}

type FakeWriter struct{}

func (f FakeWriter) Write(_ context.Context, _ model.CheckExecutionResult) error {
	return nil
}
