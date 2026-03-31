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
		executor := internal.NewCheckExecutor(
			w,
			cfg.HttpChecks[i].Interval,
			cfg.HttpChecks[i].Jitter,
		)
		checker := http.NewChecker(cfg.HttpChecks[i], executor)
		g.Go(func() error {
			return checker.Run(ctx)
		})
	}

	for i := range cfg.TCPChecks {
		executor := internal.NewCheckExecutor(w, cfg.TCPChecks[i].Interval, cfg.TCPChecks[i].Jitter)
		checker := tcp.NewChecker(cfg.TCPChecks[i], executor)
		g.Go(func() error {
			return checker.Run(ctx)
		})
	}

	for i := range cfg.GRPCChecks {
		executor := internal.NewCheckExecutor(
			w,
			cfg.GRPCChecks[i].Interval,
			cfg.GRPCChecks[i].Jitter,
		)
		checker := grpc.NewChecker(cfg.GRPCChecks[i], executor)
		g.Go(func() error {
			return checker.Run(ctx)
		})
	}

	for i := range cfg.DNSChecks {
		executor := internal.NewCheckExecutor(w, cfg.DNSChecks[i].Interval, cfg.DNSChecks[i].Jitter)
		checker := dns.NewChecker(cfg.DNSChecks[i], executor)
		g.Go(func() error {
			return checker.Run(ctx)
		})
	}

	for i := range cfg.TLSChecks {
		executor := internal.NewCheckExecutor(w, cfg.TLSChecks[i].Interval, cfg.TLSChecks[i].Jitter)
		checker := tls.NewChecker(cfg.TLSChecks[i], executor)
		g.Go(func() error {
			return checker.Run(ctx)
		})
	}

	return g.Wait()
}
