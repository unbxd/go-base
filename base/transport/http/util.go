package http

import (
	"context"
	net_http "net/http"
)

// decorateEndpoint returns endpoint.Endpoint by wrapping around HandlerFunc
func decorateEndpoint(fn HandlerFunc) Handler {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		rr, ok := req.(*net_http.Request)
		if !ok {
			return nil, ErrNotHTTPRequest
		}
		return fn(ctx, rr)
	}
}

func encapsulate(
	fn HandlerFunc,
	trs []HandlerOption,
	pats []HandlerOption,
) net_http.Handler {
	return newHandler(
		decorateEndpoint(fn),
		append(trs, pats...)...,
	)
}
