package endpoint

import (
	"context"
)

type (

	// Endpoint is a wrapper on top of kit_endpoint.Endpoint
	Endpoint func(cx context.Context, req interface{}) (res interface{}, err error)

	// Middleware is a wrapper on top of kit_endpoint.Middleware
	Middleware func(Endpoint) Endpoint
)

// NopEndpoint is no action Endpoint
func NopEndpoint(
	context.Context,
	interface{},
) (interface{}, error) {
	return struct{}{}, nil
}

// Chain is a wrapper on top of kit_endpoint.Middleware
func Chain(outer Middleware, others ...Middleware) Middleware {
	return func(next Endpoint) Endpoint {
		for i := len(others) - 1; i >= 0; i-- { // reverse
			next = others[i](next)
		}
		return outer(next)
	}
}
