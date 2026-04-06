package app

import (
	"context"
	"slices"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/pixel365/pulse/internal"
	"github.com/pixel365/pulse/internal/checker/dns"
	"github.com/pixel365/pulse/internal/checker/grpc"
	"github.com/pixel365/pulse/internal/checker/http"
	"github.com/pixel365/pulse/internal/checker/tcp"
	"github.com/pixel365/pulse/internal/checker/tls"
	"github.com/pixel365/pulse/internal/config"
	"github.com/pixel365/pulse/internal/logger"
	checksvc "github.com/pixel365/pulse/internal/services/check"
)

type manager struct {
	checkHandlerSvc  checksvc.CheckHandlerService
	logger           logger.Logger
	g                *errgroup.Group
	cfg              *config.Config
	checkersRegistry map[string]func()
	mu               sync.Mutex
}

func newManager(app *App) *manager {
	return &manager{
		cfg:              app.cfg,
		checkersRegistry: make(map[string]func()),
		checkHandlerSvc:  app.checkHandlerSvc,
		logger:           app.logger,
	}
}

func (m *manager) registerChecker(key string, cancel func()) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.checkersRegistry[key] = cancel
}

func (m *manager) run(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)
	m.g = g

	m.g.Go(func() error {
		m.listen(ctx)
		return nil
	})

	for k, v := range m.cfg.HttpChecks {
		m.runHTTPChecker(ctx, k, v)
	}

	for k, v := range m.cfg.TCPChecks {
		m.runTCPChecker(ctx, k, v)
	}

	for k, v := range m.cfg.GRPCChecks {
		m.runGRPCChecker(ctx, k, v)
	}

	for k, v := range m.cfg.DNSChecks {
		m.runDNSChecker(ctx, k, v)
	}

	for k, v := range m.cfg.TLSChecks {
		m.runTLSChecker(ctx, k, v)
	}

	return g.Wait()
}

func (m *manager) listen(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.cmp(ctx)
		}
	}
}

func (m *manager) runHTTPChecker(
	ctx context.Context,
	key string,
	obj config.TypedCheck[config.HttpSpec],
) {
	key = checkKey(config.HTTP, key)
	m.mu.Lock()

	if !obj.Enabled {
		if cancel, found := m.checkersRegistry[key]; found {
			cancel()
			delete(m.checkersRegistry, key)
		}
		m.mu.Unlock()
		return
	}

	if _, found := m.checkersRegistry[key]; found {
		m.mu.Unlock()
		return
	}

	m.mu.Unlock()

	executor := internal.NewCheckExecutor(
		m.checkHandlerSvc,
		obj.CheckFields,
		m.logger,
	)
	checker := http.NewChecker(obj, executor)

	cancelCtx, cancel := context.WithCancel(ctx)
	m.registerChecker(key, func() { cancel() })

	m.g.Go(func() error {
		return checker.Check(cancelCtx)
	})
}

func (m *manager) runTCPChecker(
	ctx context.Context,
	key string,
	obj config.TypedCheck[config.TCPSpec],
) {
	key = checkKey(config.TCP, key)
	m.mu.Lock()

	if !obj.Enabled {
		if cancel, found := m.checkersRegistry[key]; found {
			cancel()
			delete(m.checkersRegistry, key)
		}
		m.mu.Unlock()
		return
	}

	if _, found := m.checkersRegistry[key]; found {
		m.mu.Unlock()
		return
	}

	m.mu.Unlock()

	executor := internal.NewCheckExecutor(
		m.checkHandlerSvc,
		obj.CheckFields,
		m.logger,
	)
	checker := tcp.NewChecker(obj, executor)

	cancelCtx, cancel := context.WithCancel(ctx)
	m.registerChecker(key, func() { cancel() })

	m.g.Go(func() error {
		return checker.Check(cancelCtx)
	})
}

func (m *manager) runGRPCChecker(
	ctx context.Context,
	key string,
	obj config.TypedCheck[config.GRPCSpec],
) {
	key = checkKey(config.GRPC, key)
	m.mu.Lock()

	if !obj.Enabled {
		if cancel, found := m.checkersRegistry[key]; found {
			cancel()
			delete(m.checkersRegistry, key)
		}
		m.mu.Unlock()
		return
	}

	if _, found := m.checkersRegistry[key]; found {
		m.mu.Unlock()
		return
	}

	m.mu.Unlock()

	executor := internal.NewCheckExecutor(
		m.checkHandlerSvc,
		obj.CheckFields,
		m.logger,
	)
	checker := grpc.NewChecker(obj, executor)

	cancelCtx, cancel := context.WithCancel(ctx)
	m.registerChecker(key, func() { cancel() })

	m.g.Go(func() error {
		return checker.Check(cancelCtx)
	})
}

func (m *manager) runDNSChecker(
	ctx context.Context,
	key string,
	obj config.TypedCheck[config.DNSSpec],
) {
	key = checkKey(config.DNS, key)
	m.mu.Lock()

	if !obj.Enabled {
		if cancel, found := m.checkersRegistry[key]; found {
			cancel()
			delete(m.checkersRegistry, key)
		}
		m.mu.Unlock()
		return
	}

	if _, found := m.checkersRegistry[key]; found {
		m.mu.Unlock()
		return
	}

	m.mu.Unlock()

	executor := internal.NewCheckExecutor(
		m.checkHandlerSvc,
		obj.CheckFields,
		m.logger,
	)
	checker := dns.NewChecker(obj, executor)

	cancelCtx, cancel := context.WithCancel(ctx)
	m.registerChecker(key, func() { cancel() })

	m.g.Go(func() error {
		return checker.Check(cancelCtx)
	})
}

func (m *manager) runTLSChecker(
	ctx context.Context,
	key string,
	obj config.TypedCheck[config.TLSSpec],
) {
	key = checkKey(config.TLS, key)
	m.mu.Lock()

	if !obj.Enabled {
		if cancel, found := m.checkersRegistry[key]; found {
			cancel()
			delete(m.checkersRegistry, key)
		}
		m.mu.Unlock()
		return
	}

	if _, found := m.checkersRegistry[key]; found {
		m.mu.Unlock()
		return
	}

	m.mu.Unlock()

	executor := internal.NewCheckExecutor(
		m.checkHandlerSvc,
		obj.CheckFields,
		m.logger,
	)
	checker := tls.NewChecker(obj, executor)

	cancelCtx, cancel := context.WithCancel(ctx)
	m.registerChecker(key, func() { cancel() })

	m.g.Go(func() error {
		return checker.Check(cancelCtx)
	})
}

func (m *manager) cmp(ctx context.Context) {
	cfg, err := config.Load()
	if err != nil {
		m.logger.Error(ctx, "invalid config", "error", err)
		return
	}

	m.mu.Lock()
	oldCfg := m.cfg
	m.mu.Unlock()

	reconcileChecks(m, config.HTTP, oldCfg.HttpChecks, cfg.HttpChecks, sameHTTPCheck)
	reconcileChecks(m, config.TCP, oldCfg.TCPChecks, cfg.TCPChecks, sameTCPCheck)
	reconcileChecks(m, config.GRPC, oldCfg.GRPCChecks, cfg.GRPCChecks, sameGRPCCheck)
	reconcileChecks(m, config.DNS, oldCfg.DNSChecks, cfg.DNSChecks, sameDNSCheck)
	reconcileChecks(m, config.TLS, oldCfg.TLSChecks, cfg.TLSChecks, sameTLSCheck)

	for k, v := range cfg.HttpChecks {
		m.runHTTPChecker(ctx, k, v)
	}

	for k, v := range cfg.TCPChecks {
		m.runTCPChecker(ctx, k, v)
	}

	for k, v := range cfg.GRPCChecks {
		m.runGRPCChecker(ctx, k, v)
	}

	for k, v := range cfg.DNSChecks {
		m.runDNSChecker(ctx, k, v)
	}

	for k, v := range cfg.TLSChecks {
		m.runTLSChecker(ctx, k, v)
	}

	m.mu.Lock()
	m.cfg = cfg
	m.mu.Unlock()
}

func reconcileChecks[T any](
	m *manager,
	checkType config.CheckType,
	oldChecks map[string]config.TypedCheck[T],
	newChecks map[string]config.TypedCheck[T],
	same func(config.TypedCheck[T], config.TypedCheck[T]) bool,
) {
	for key, oldCheck := range oldChecks {
		newCheck, found := newChecks[key]
		if !found || !newCheck.Enabled || !same(oldCheck, newCheck) {
			m.cancelChecker(checkKey(checkType, key))
		}
	}
}

func (m *manager) cancelChecker(key string) {
	m.mu.Lock()
	cancel, found := m.checkersRegistry[key]
	if found {
		delete(m.checkersRegistry, key)
	}
	m.mu.Unlock()

	if found {
		cancel()
	}
}

func checkKey(checkType config.CheckType, key string) string {
	return string(checkType) + ":" + key
}

func sameHTTPCheck(a, b config.TypedCheck[config.HttpSpec]) bool {
	return sameCheckFields(a.CheckFields, b.CheckFields) && (&a.Spec).Cmp(&b.Spec)
}

func sameTCPCheck(a, b config.TypedCheck[config.TCPSpec]) bool {
	return sameCheckFields(a.CheckFields, b.CheckFields) && (&a.Spec).Cmp(&b.Spec)
}

func sameGRPCCheck(a, b config.TypedCheck[config.GRPCSpec]) bool {
	return sameCheckFields(a.CheckFields, b.CheckFields) && (&a.Spec).Cmp(&b.Spec)
}

func sameDNSCheck(a, b config.TypedCheck[config.DNSSpec]) bool {
	return sameCheckFields(a.CheckFields, b.CheckFields) && (&a.Spec).Cmp(&b.Spec)
}

func sameTLSCheck(a, b config.TypedCheck[config.TLSSpec]) bool {
	return sameCheckFields(a.CheckFields, b.CheckFields) && (&a.Spec).Cmp(&b.Spec)
}

func sameCheckFields(a, b config.CheckFields) bool {
	return a.ID == b.ID &&
		a.Name == b.Name &&
		a.Service == b.Service &&
		a.Type == b.Type &&
		slices.Equal(a.Tags, b.Tags) &&
		a.Timeout == b.Timeout &&
		a.Jitter == b.Jitter &&
		a.Retries == b.Retries &&
		a.FailureThreshold == b.FailureThreshold &&
		a.SuccessThreshold == b.SuccessThreshold &&
		a.Interval == b.Interval &&
		a.Enabled == b.Enabled
}
