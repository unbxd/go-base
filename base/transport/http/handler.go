package http

import (
	net_http "net/http"

	"context"

	kit_endpoint "github.com/go-kit/kit/endpoint"
	kit_http "github.com/go-kit/kit/transport/http"
	"github.com/unbxd/go-base/base/endpoint"
)

type (
	// HandlerFunc defines the default wrapper for request/response handling
	// in http transport.
	// By default we can't support the default net_http.handler interface as
	// the `ServeHTTP(r *http.Request, rw http.ResponseWriter)` func has
	// rw built in which overlaps our abstraction
	// This here instead gives us the ability to expose this as endpoint.Endpoint
	// and use it in the go-kit chain
	HandlerFunc func(ctx context.Context, req *net_http.Request) (*net_http.Response, error)

	// Handler is wrapper on top of endpoint.Endpoint
	Handler endpoint.Endpoint

	// handler is wrapper on top of kit_http.Server
	handler struct {
		*kit_http.Server

		endpoint endpoint.Endpoint

		encoder      Encoder
		decoder      Decoder
		errorEncoder ErrorEncoder
		errorhandler ErrorHandler

		options []kit_http.ServerOption
	}

	// HandlerOption provides ways to modify the handler
	HandlerOption func(*handler)
)

// HandlerWithEncoder returns a request handler with customer encoder function
func HandlerWithEncoder(fn Encoder) HandlerOption {
	return func(h *handler) { h.encoder = fn }
}

// HandlerWithDecoder returns a request handler with a customer decoer function
func HandlerWithDecoder(fn Decoder) HandlerOption {
	return func(h *handler) { h.decoder = fn }
}

// HandlerWithErrorEncoder returns a request handler with a customer error
// encoder function
func HandlerWithErrorEncoder(fn ErrorEncoder) HandlerOption {
	return func(h *handler) {
		h.errorEncoder = fn
		h.options = append(h.options, kit_http.ServerErrorEncoder(
			kit_http.ErrorEncoder(fn),
		))
	}
}

// HandlerWithErrorhandler returns a request handler with a custom error handler
func HandlerWithErrorhandler(fn ErrorHandler) HandlerOption {
	return func(h *handler) {
		h.errorhandler = fn
		h.options = append(h.options, kit_http.ServerErrorHandler(fn))
	}
}

// newhandler returns a new handler
func newHandler(fn Handler, options ...HandlerOption) *handler {
	hn := &handler{
		encoder:      nil,
		decoder:      nil,
		errorEncoder: nil,
		errorhandler: nil,
		options: []kit_http.ServerOption{
			// default server options here
			kit_http.ServerBefore(populateRequestContext),
		},
	}

	for _, o := range options {
		o(hn)
	}

	if hn.encoder == nil {
		// Todo: throw a warn
		hn.encoder = newDefaultEncoder()
	}

	if hn.decoder == nil {
		// Todo: throw a warn
		hn.decoder = newDefaultDecoder()
	}

	hn.Server = kit_http.NewServer(
		kit_endpoint.Endpoint(fn),
		kit_http.DecodeRequestFunc(hn.decoder),
		kit_http.EncodeResponseFunc(hn.encoder),
		hn.options...,
	)

	return hn
}
