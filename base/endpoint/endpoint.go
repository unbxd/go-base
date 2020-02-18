package endpoint

import (
	"context"

	kit_endpoint "github.com/go-kit/kit/endpoint"
)

type (

	// Endpoint is a wrapper on top of kit_endpoint.Endpoint
	Endpoint kit_endpoint.Endpoint

	// Middleware is a wrapper on top of kit_endpoint.Middleware
	Middleware kit_endpoint.Middleware
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
	o := []kit_endpoint.Middleware{}

	for _, oth := range others {
		o = append(o, kit_endpoint.Middleware(oth))
	}

	return Middleware(kit_endpoint.Chain(
		kit_endpoint.Middleware(outer),
		o...,
	))
}
