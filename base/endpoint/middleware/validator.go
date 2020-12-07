package middleware

import (
	"context"
	"github.com/apoorvprecisely/go-base/base/endpoint"
	"github.com/apoorvprecisely/go-base/base/validator"
)

func ValidatorMw(v validator.Validator) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(cx context.Context, req interface{}) (interface{}, error) {
			err := v.Validate(cx, req)
			if err != nil {
				return nil, err
			}
			return next(cx, req)
		}
	}
}
