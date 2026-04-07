package config

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pixel365/pulse/internal/logger"
)

var _ ConfigProvider = (*CurrentProvider)(nil)

type ConfigProvider interface {
	Current() *Config
	Start(context.Context)
}

type CurrentProvider struct {
	current atomic.Pointer[Config]
	log     logger.Logger
	once    sync.Once
}

func NewCurrentProvider(cfg *Config, log logger.Logger) *CurrentProvider {
	p := &CurrentProvider{log: log}
	p.current.Store(cfg)

	return p
}

func (p *CurrentProvider) Start(ctx context.Context) {
	p.once.Do(func() {
		go p.listen(ctx)
	})
}

func (p *CurrentProvider) Current() *Config {
	return p.current.Load()
}

func (p *CurrentProvider) listen(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c, err := Load()
			if err == nil {
				p.reload(c)
			} else {
				p.log.Error(ctx, "failed to reload config", "error", err)
			}
		}
	}
}

func (p *CurrentProvider) reload(cfg *Config) {
	p.current.Store(cfg)
}
