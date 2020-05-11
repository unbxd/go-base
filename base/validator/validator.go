package validator

import "context"

type (
	Validator interface {
		Validate(cx context.Context, req interface{}) error
	}
)
