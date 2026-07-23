package request

import "github.com/go-playground/validator/v10"

var validate = validator.New()

func IsValid[T any](payload T) error {
	return validate.Struct(payload)
}
