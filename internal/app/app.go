package app

import (
	"context"

	"github.com/pixel365/pulse/internal/logger"

	checksvc "github.com/pixel365/pulse/internal/services/check"

	"github.com/pixel365/pulse/internal"

	"github.com/pixel365/pulse/internal/config"
)

var _ internal.Runner = (*App)(nil)

type App struct {
	cfg             *config.Config
	checkHandlerSvc checksvc.CheckHandlerService
	logger          logger.Logger
}

func NewApp(
	cfg *config.Config,
	logger logger.Logger,
	checkSvc checksvc.CheckHandlerService,
) *App {
	return &App{
		cfg:             cfg,
		checkHandlerSvc: checkSvc,
		logger:          logger,
	}
}

func (a *App) Run(ctx context.Context) error {
	m := newManager(a)
	return m.run(ctx)
}
