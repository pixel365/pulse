package config

import "sync/atomic"

type ConfigProvider interface {
	Current() *Config
}

type CurrentProvider struct {
	current atomic.Pointer[Config]
}

func NewCurrentProvider(cfg *Config) *CurrentProvider {
	p := &CurrentProvider{}
	p.current.Store(cfg)

	return p
}

func (p *CurrentProvider) Current() *Config {
	return p.current.Load()
}

func (p *CurrentProvider) Reload(cfg *Config) {
	p.current.Store(cfg)
}
