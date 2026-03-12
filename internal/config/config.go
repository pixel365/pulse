package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
)

type Config struct {
	Services   map[string]Service
	HttpChecks map[string]TypedCheck[HttpSpec]
	TCPChecks  map[string]TypedCheck[TCPSpec]
	GRPCChecks map[string]TypedCheck[GRPCSpec]
	DNSChecks  map[string]TypedCheck[DNSSpec]
	TLSChecks  map[string]TypedCheck[TLSSpec]
	dir        string
}

func MustLoad() *Config {
	dir := os.Getenv("CONFIG_DIR")
	if dir == "" {
		panic("CONFIG_DIR is not set")
	}

	cfg := &Config{
		dir:        filepath.Clean(dir),
		Services:   make(map[string]Service),
		HttpChecks: make(map[string]TypedCheck[HttpSpec]),
		TCPChecks:  make(map[string]TypedCheck[TCPSpec]),
		GRPCChecks: make(map[string]TypedCheck[GRPCSpec]),
		DNSChecks:  make(map[string]TypedCheck[DNSSpec]),
		TLSChecks:  make(map[string]TypedCheck[TLSSpec]),
	}
	if err := cfg.load(); err != nil {
		panic(err)
	}

	return cfg
}

func (c *Config) load() error {
	servicesFilePath := filepath.Join(c.dir, "services.yaml")
	servicesFilePath = filepath.Clean(servicesFilePath)

	file, err := os.ReadFile(servicesFilePath)
	if err != nil {
		return fmt.Errorf("failed to read services file: %w", err)
	}

	var services ServiceSet
	if err := yaml.Unmarshal(file, &services); err != nil {
		return fmt.Errorf("failed to unmarshal services: %w", err)
	}

	if err = validateStruct(services); err != nil {
		return fmt.Errorf("failed to validate services: %w", err)
	}

	for i := range services.Services {
		svc := services.Services[i]
		if _, found := c.Services[svc.ID]; found {
			return fmt.Errorf("duplicate service id: %s", svc.Name)
		}
		c.Services[svc.ID] = svc
	}

	if err = c.readChecks(); err != nil {
		return fmt.Errorf("failed to read checks: %w", err)
	}

	return nil
}

func (c *Config) readChecks() error {
	dir := filepath.Join(c.dir, "checks")
	dir = filepath.Clean(dir)

	entries, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || entry.Name() == "services.yaml" ||
			!strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		filePath := filepath.Join(dir, entry.Name())
		filePath = filepath.Clean(filePath)
		file, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}

		var checkSet CheckSet
		if err = yaml.Unmarshal(file, &checkSet); err != nil {
			return fmt.Errorf(
				"failed to unmarshal check config file %s: %w",
				entry.Name(),
				err,
			)
		}

		if err = validateStruct(checkSet); err != nil {
			return fmt.Errorf("failed to validate check config file %s: %w", entry.Name(), err)
		}

		if err = c.handleChecks(checkSet, entry.Name()); err != nil {
			return err
		}
	}

	return nil
}

func (c *Config) handleChecks(set CheckSet, filename string) error {
	for i := range set.Checks {
		if _, found := c.Services[set.Checks[i].Service]; !found {
			return fmt.Errorf("service %s not found", set.Checks[i].Service)
		}

		var err error
		switch set.Checks[i].Type {
		case HTTP:
			err = appendTypedCheck[HttpSpec](c.HttpChecks, set.Checks[i])
		case TCP:
			err = appendTypedCheck[TCPSpec](c.TCPChecks, set.Checks[i])
		case GRPC:
			err = appendTypedCheck[GRPCSpec](c.GRPCChecks, set.Checks[i])
		case DNS:
			err = appendTypedCheck[DNSSpec](c.DNSChecks, set.Checks[i])
		case TLS:
			err = appendTypedCheck[TLSSpec](c.TLSChecks, set.Checks[i])
		}

		if err != nil {
			return fmt.Errorf("failed to read check config file %s: %w", filename, err)
		}
	}

	return nil
}

func decodeTypedCheck[T any](raw Check) (TypedCheck[T], error) {
	var typedCheck TypedCheck[T]

	enc, err := yaml.Marshal(raw)
	if err != nil {
		return typedCheck, fmt.Errorf("failed to marshal check spec: %w", err)
	}

	if err = yaml.Unmarshal(enc, &typedCheck); err != nil {
		return typedCheck, fmt.Errorf("failed to unmarshal check spec: %w", err)
	}

	if err = validateStruct(typedCheck); err != nil {
		return typedCheck, fmt.Errorf("failed to validate check spec: %w", err)
	}

	return typedCheck, nil
}

func appendTypedCheck[T any](dst map[string]TypedCheck[T], raw Check) error {
	if _, found := dst[raw.Name]; found {
		return fmt.Errorf("duplicate check name: %s", raw.Name)
	}

	typedCheck, err := decodeTypedCheck[T](raw)
	if err != nil {
		return err
	}

	dst[raw.Name] = typedCheck

	return nil
}
