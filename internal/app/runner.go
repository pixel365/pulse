package app

import (
	"context"

	"golang.org/x/sync/errgroup"

	"github.com/pixel365/pulse/internal/checker/dns"
	"github.com/pixel365/pulse/internal/checker/grpc"
	"github.com/pixel365/pulse/internal/checker/http"
	"github.com/pixel365/pulse/internal/checker/tcp"
	"github.com/pixel365/pulse/internal/checker/tls"
	"github.com/pixel365/pulse/internal/config"
)

var _ Runner = (*App)(nil)

type Runner interface {
	Run(context.Context, *config.Config) error
}

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Run(ctx context.Context, cfg *config.Config) error {
	g, ctx := errgroup.WithContext(ctx)

	for i := range cfg.HttpChecks {
		checker := http.NewChecker(cfg.HttpChecks[i])
		g.Go(func() error {
			return checker.Run(ctx)
		})
	}

	for i := range cfg.TCPChecks {
		checker := tcp.NewChecker(cfg.TCPChecks[i])
		g.Go(func() error {
			return checker.Run(ctx)
		})
	}

	for i := range cfg.GRPCChecks {
		checker := grpc.NewChecker(cfg.GRPCChecks[i])
		g.Go(func() error {
			return checker.Run(ctx)
		})
	}

	for i := range cfg.DNSChecks {
		checker := dns.NewChecker(cfg.DNSChecks[i])
		g.Go(func() error {
			return checker.Run(ctx)
		})
	}

	for i := range cfg.TLSChecks {
		checker := tls.NewChecker(cfg.TLSChecks[i])
		g.Go(func() error {
			return checker.Run(ctx)
		})
	}

	return g.Wait()
}
