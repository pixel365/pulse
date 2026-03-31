package app

import (
	"context"

	"golang.org/x/sync/errgroup"

	"github.com/pixel365/pulse/internal"

	"github.com/pixel365/pulse/internal/checker/dns"
	"github.com/pixel365/pulse/internal/checker/grpc"
	"github.com/pixel365/pulse/internal/checker/http"
	"github.com/pixel365/pulse/internal/checker/tcp"
	"github.com/pixel365/pulse/internal/checker/tls"
	"github.com/pixel365/pulse/internal/config"
)

var _ internal.Runner = (*App)(nil)

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Run(ctx context.Context, cfg *config.Config) error {
	g, ctx := errgroup.WithContext(ctx)

	w := internal.FakeWriter{}

	for i := range cfg.HttpChecks {
		executor := internal.NewCheckExecutor(w, cfg.HttpChecks[i].CheckFields)
		checker := http.NewChecker(cfg.HttpChecks[i], executor)
		g.Go(func() error {
			return checker.Run(ctx)
		})
	}

	for i := range cfg.TCPChecks {
		executor := internal.NewCheckExecutor(w, cfg.TCPChecks[i].CheckFields)
		checker := tcp.NewChecker(cfg.TCPChecks[i], executor)
		g.Go(func() error {
			return checker.Run(ctx)
		})
	}

	for i := range cfg.GRPCChecks {
		executor := internal.NewCheckExecutor(w, cfg.GRPCChecks[i].CheckFields)
		checker := grpc.NewChecker(cfg.GRPCChecks[i], executor)
		g.Go(func() error {
			return checker.Run(ctx)
		})
	}

	for i := range cfg.DNSChecks {
		executor := internal.NewCheckExecutor(w, cfg.DNSChecks[i].CheckFields)
		checker := dns.NewChecker(cfg.DNSChecks[i], executor)
		g.Go(func() error {
			return checker.Run(ctx)
		})
	}

	for i := range cfg.TLSChecks {
		executor := internal.NewCheckExecutor(w, cfg.TLSChecks[i].CheckFields)
		checker := tls.NewChecker(cfg.TLSChecks[i], executor)
		g.Go(func() error {
			return checker.Run(ctx)
		})
	}

	return g.Wait()
}
