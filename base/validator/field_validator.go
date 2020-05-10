package validator

import (
	"context"
	"github.com/go-playground/validator/v10"
	"github.com/unbxd/go-base/base/endpoint"
)

type (
	FieldValidationError error

	FieldValidatorOption func(*FieldValidator)

	FieldValidator struct {
		validate *validator.Validate
		tag      string
	}
)

func (f *FieldValidator) Validate(cx context.Context, req interface{}) error {
	err := f.validate.Struct(req)
	if err != nil {
		return FieldValidationError(err)
	}
	return err
}

func WithTag(tag string) FieldValidatorOption {
	return func(f *FieldValidator) {
		f.tag = tag
	}
}

func NewFieldValidator(opts ...FieldValidatorOption) Validator {
	f := &FieldValidator{
		validate: validator.New(),
	}

	for _, opt := range opts {
		opt(f)
	}

	if f.tag != "" {
		f.validate.SetTagName(f.tag)
	}
	return f
}

func NewFieldValidatorMw(opts ...FieldValidatorOption) endpoint.Middleware {
	return MiddleWare(NewFieldValidator(opts...))
}
