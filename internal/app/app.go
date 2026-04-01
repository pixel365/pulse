package app

import (
	"context"

	"golang.org/x/sync/errgroup"

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
}

func NewApp(cfg *config.Config, checkSvc checksvc.CheckHandlerService) *App {
	return &App{
		cfg:             cfg,
		checkHandlerSvc: checkSvc,
	}
}

func (a *App) Run(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)

	for i := range a.cfg.HttpChecks {
		executor := internal.NewCheckExecutor(a.checkHandlerSvc, a.cfg.HttpChecks[i].CheckFields)
		checker := http.NewChecker(a.cfg.HttpChecks[i], executor)
		g.Go(func() error {
			return checker.Check(ctx)
		})
	}

	for i := range a.cfg.TCPChecks {
		executor := internal.NewCheckExecutor(a.checkHandlerSvc, a.cfg.TCPChecks[i].CheckFields)
		checker := tcp.NewChecker(a.cfg.TCPChecks[i], executor)
		g.Go(func() error {
			return checker.Check(ctx)
		})
	}

	for i := range a.cfg.GRPCChecks {
		executor := internal.NewCheckExecutor(a.checkHandlerSvc, a.cfg.GRPCChecks[i].CheckFields)
		checker := grpc.NewChecker(a.cfg.GRPCChecks[i], executor)
		g.Go(func() error {
			return checker.Check(ctx)
		})
	}

	for i := range a.cfg.DNSChecks {
		executor := internal.NewCheckExecutor(a.checkHandlerSvc, a.cfg.DNSChecks[i].CheckFields)
		checker := dns.NewChecker(a.cfg.DNSChecks[i], executor)
		g.Go(func() error {
			return checker.Check(ctx)
		})
	}

	for i := range a.cfg.TLSChecks {
		executor := internal.NewCheckExecutor(a.checkHandlerSvc, a.cfg.TLSChecks[i].CheckFields)
		checker := tls.NewChecker(a.cfg.TLSChecks[i], executor)
		g.Go(func() error {
			return checker.Check(ctx)
		})
	}

	return g.Wait()
}
