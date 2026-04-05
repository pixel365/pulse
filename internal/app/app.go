package app

import (
	"context"

	"golang.org/x/sync/errgroup"

	"github.com/pixel365/pulse/internal/logger"

	checksvc "github.com/pixel365/pulse/internal/services/check"

	"github.com/pixel365/pulse/internal"

	"github.com/pixel365/pulse/internal/checker/dns"
	"github.com/pixel365/pulse/internal/checker/grpc"
	"github.com/pixel365/pulse/internal/checker/http"
	"github.com/pixel365/pulse/internal/checker/tcp"
	"github.com/pixel365/pulse/internal/checker/tls"
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
	g, ctx := errgroup.WithContext(ctx)

	for i := range a.cfg.HttpChecks {
		if !a.cfg.HttpChecks[i].Enabled {
			continue
		}

		executor := internal.NewCheckExecutor(
			a.checkHandlerSvc,
			a.cfg.HttpChecks[i].CheckFields,
			a.logger,
		)
		checker := http.NewChecker(a.cfg.HttpChecks[i], executor)
		g.Go(func() error {
			return checker.Check(ctx)
		})
	}

	for i := range a.cfg.TCPChecks {
		if !a.cfg.TCPChecks[i].Enabled {
			continue
		}

		executor := internal.NewCheckExecutor(
			a.checkHandlerSvc,
			a.cfg.TCPChecks[i].CheckFields,
			a.logger,
		)
		checker := tcp.NewChecker(a.cfg.TCPChecks[i], executor)
		g.Go(func() error {
			return checker.Check(ctx)
		})
	}

	for i := range a.cfg.GRPCChecks {
		if !a.cfg.GRPCChecks[i].Enabled {
			continue
		}

		executor := internal.NewCheckExecutor(
			a.checkHandlerSvc,
			a.cfg.GRPCChecks[i].CheckFields,
			a.logger,
		)
		checker := grpc.NewChecker(a.cfg.GRPCChecks[i], executor)
		g.Go(func() error {
			return checker.Check(ctx)
		})
	}

	for i := range a.cfg.DNSChecks {
		if !a.cfg.DNSChecks[i].Enabled {
			continue
		}

		executor := internal.NewCheckExecutor(
			a.checkHandlerSvc,
			a.cfg.DNSChecks[i].CheckFields,
			a.logger,
		)
		checker := dns.NewChecker(a.cfg.DNSChecks[i], executor)
		g.Go(func() error {
			return checker.Check(ctx)
		})
	}

	for i := range a.cfg.TLSChecks {
		if !a.cfg.TLSChecks[i].Enabled {
			continue
		}

		executor := internal.NewCheckExecutor(
			a.checkHandlerSvc,
			a.cfg.TLSChecks[i].CheckFields,
			a.logger,
		)
		checker := tls.NewChecker(a.cfg.TLSChecks[i], executor)
		g.Go(func() error {
			return checker.Check(ctx)
		})
	}

	return g.Wait()
}
