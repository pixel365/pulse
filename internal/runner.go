package internal

import (
	"context"

	"github.com/pixel365/pulse/internal/config"
)

type Runner interface {
	Run(context.Context, *config.Config) error
}
