package config

import (
	"errors"
	"fmt"
	"os"
	"regexp"

	"github.com/go-playground/validator/v10"
)

var (
	validate      = validator.New()
	envRefPattern = regexp.MustCompile(`\$\{([A-Z][A-Z0-9_]*)}`)
)

func init() {
	if err := validate.RegisterValidation("env_url", validateEnvURL); err != nil {
		panic(err)
	}

	if err := validate.RegisterValidation("env_host", validateEnvHost); err != nil {
		panic(err)
	}

	if err := validate.RegisterValidation("env_http_method", validateEnvHTTPMethod); err != nil {
		panic(err)
	}
}

type Validator interface {
	Validate() error
}

func validateStruct(s any) error {
	if err := validate.Struct(s); err != nil {
		return err
	}

	if v, ok := s.(Validator); ok {
		return v.Validate()
	}

	return nil
}

func validateMap(m map[string]string) error {
	for k, v := range m {
		if k == "" {
			return errors.New("map key cannot be empty")
		}

		if v == "" {
			return fmt.Errorf("map value for key %s cannot be empty", k)
		}

		if err := validateEnvRefs(v); err != nil {
			return fmt.Errorf(
				"invalid environment variable reference in map value for key %s: %w",
				k,
				err,
			)
		}
	}

	return nil
}

func validateString(s string) error {
	if s == "" {
		return errors.New("string cannot be empty")
	}

	return validateEnvRefs(s)
}

func validateEnvRefs(value string) error {
	for _, match := range envRefPattern.FindAllStringSubmatch(value, -1) {
		name := match[1]
		if _, ok := os.LookupEnv(name); !ok {
			return fmt.Errorf("environment variable %s is not set", name)
		}
	}

	return nil
}

func validateEnvURL(fl validator.FieldLevel) bool {
	value, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}

	if err := validateEnvRefs(value); err != nil {
		return false
	}

	return validate.Var(parseEnv(value), "url") == nil
}

func validateEnvHost(fl validator.FieldLevel) bool {
	value, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}

	if err := validateEnvRefs(value); err != nil {
		return false
	}

	return validate.Var(parseEnv(value), "hostname_rfc1123|ip") == nil
}

func validateEnvHTTPMethod(fl validator.FieldLevel) bool {
	value, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}

	if err := validateEnvRefs(value); err != nil {
		return false
	}

	return validate.Var(parseEnv(value), "oneof=GET POST") == nil
}
