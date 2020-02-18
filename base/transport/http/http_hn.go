package http

import (
	net_http "net/http"

	kit_http "github.com/go-kit/kit/transport/http"

	"github.com/go-kit/kit/endpoint"

	"context"

	"github.com/oxtoacart/bpool"
)

// HandlerFunc defines the default wrapper for request/response handling
// in http transport.
// By default we can't support the default net_http.Handler interface as
// the `ServeHTTP(r *http.Request, rw http.ResponseWriter)` func has
// rw built in which overlaps our abstraction
// This here instead gives us the ability to expose this as endpoint.Endpoint
// and use it in the go-kit chain
type HandlerFunc func(ctx context.Context, req *net_http.Request) (*net_http.Response, error)

// Endpoint returns endpoint.Endpoint by wrapping around HandlerFunc
func Endpoint(fn HandlerFunc) endpoint.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		rr, ok := req.(*net_http.Request)
		if !ok {
			return nil, ErrNotHTTPRequest
		}

		return fn(ctx, rr)
	}
}

func newDefaultDecoder() kit_http.DecodeRequestFunc {
	return func(ctx context.Context, r *net_http.Request) (interface{}, error) {
		return r, nil
	}
}

func newDefaultEncoder() kit_http.EncodeResponseFunc {
	bufferPool := bpool.NewBytePool(100, 1000000)

	return func(ctx context.Context, rw net_http.ResponseWriter, res interface{}) (err error) {
		rr, ok := res.(*net_http.Response)
		if !ok {
			return ErrNotHTTPResponse
		}

		if res == nil {
			rw.WriteHeader(net_http.StatusNoContent)
			return
		}

		copyHeader(rw.Header(), rr.Header)

		switch {
		case rr.StatusCode == 0:
			rw.WriteHeader(net_http.StatusOK)
		case rr.StatusCode > 0:
			rw.WriteHeader(rr.StatusCode)
		default:
			panic("status code should be non-negative")
		}

		return copyResponse(bufferPool, rw, rr.Body, flushInterval(rr))
	}
}

func encapsulate(
	fn HandlerFunc,
	defaultopts []kit_http.ServerOption,
	handleropts []HandlerOption,
) net_http.Handler {
	options := defaultopts

	for _, ho := range handleropts {
		options = append(options, kit_http.ServerOption(ho))
	}

	return kit_http.NewServer(
		Endpoint(fn),
		newDefaultDecoder(),
		newDefaultEncoder(),
		options...,
	)
}
