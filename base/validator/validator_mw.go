package validator

import (
	"context"
	"github.com/unbxd/go-base/base/endpoint"
)

func MiddleWare(v Validator) endpoint.Middleware {
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
