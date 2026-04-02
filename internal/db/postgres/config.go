package postgres

import (
	"fmt"
	"net/url"
	"os"
)

type ConfigOption func(*Config)

type Config struct {
	User     string
	Password string
	Host     string
	Port     string
	Database string
	SSLMode  string
}

func (c Config) Validate() error {
	if c.User == "" {
		return fmt.Errorf("user is empty")
	}

	if c.Password == "" {
		return fmt.Errorf("password is empty")
	}

	if c.Host == "" {
		return fmt.Errorf("host is empty")
	}

	if c.Port == "" {
		return fmt.Errorf("port is empty")
	}

	if c.Database == "" {
		return fmt.Errorf("database is empty")
	}

	return nil
}

func (c Config) DSN() string {
	user := url.UserPassword(c.User, c.Password)
	return fmt.Sprintf(
		"postgres://%s@%s:%s/%s?sslmode=%s",
		user.String(), c.Host, c.Port, c.Database, c.SSLMode,
	)
}

func NewConfig(opts ...ConfigOption) Config {
	cfg := &Config{}
	for _, opt := range opts {
		opt(cfg)
	}
	return *cfg
}

func NewConfigFromEnv() Config {
	return NewConfig(
		WithUser(os.Getenv("POSTGRES_USER")),
		WithPassword(os.Getenv("POSTGRES_PASSWORD")),
		WithHost(os.Getenv("POSTGRES_HOST")),
		WithPort(os.Getenv("POSTGRES_PORT")),
		WithDatabase(os.Getenv("POSTGRES_DB")),
		WithSSLMode(os.Getenv("POSTGRES_SSL_MODE")),
	)
}

func WithUser(user string) ConfigOption {
	return func(cfg *Config) {
		cfg.User = user
	}
}

func WithPassword(password string) ConfigOption {
	return func(cfg *Config) {
		cfg.Password = password
	}
}

func WithHost(host string) ConfigOption {
	return func(cfg *Config) {
		cfg.Host = host
	}
}

func WithPort(port string) ConfigOption {
	return func(cfg *Config) {
		cfg.Port = port
	}
}

func WithDatabase(database string) ConfigOption {
	return func(cfg *Config) {
		cfg.Database = database
	}
}

func WithSSLMode(sslMode string) ConfigOption {
	return func(cfg *Config) {
		cfg.SSLMode = sslMode
	}
}
