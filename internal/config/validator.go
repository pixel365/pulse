package config

import "github.com/go-playground/validator/v10"

var (
	validate *validator.Validate
)

func init() {
	validate = validator.New()
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
