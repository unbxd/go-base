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

	// Middleware defines the middleware for request
	Middleware endpoint.Middleware

	// handler is wrapper on top of kit_http.Server
	handler struct {
		*kit_http.Server

		endpoint endpoint.Endpoint

		encoder      Encoder
		decoder      Decoder
		errorEncoder ErrorEncoder
		errorhandler ErrorHandler
		middlewares  []Middleware

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

// HandlerWithMiddleware sets middleware for request
func HandlerWithMiddleware(fn Middleware) HandlerOption {
	return func(h *handler) {
		h.middlewares = append(h.middlewares, fn)
	}
}

// HandlerWithEndpointMiddleware provides an ability to add a
// middleware of the base type
func HandlerWithEndpointMiddleware(fn endpoint.Middleware) HandlerOption {
	return func(h *handler) {
		h.middlewares = append(h.middlewares, Middleware(fn))
	}
}

// NoopMiddleware is middleware that does nothing
// It is returned if a given middleware is not enabled
func NoopMiddleware(next endpoint.Endpoint) endpoint.Endpoint {
	return func(
		ctx context.Context,
		req interface{},
	) (interface{}, error) {
		return next(ctx, req)
	}
}

// Wrap wraps around middleware
func Wrap(hn Handler, mws ...Middleware) Handler {
	var emws []endpoint.Middleware

	for _, mw := range mws {
		emws = append(emws, endpoint.Middleware(mw))
	}

	newmw := endpoint.Chain(
		NoopMiddleware,
		emws...,
	)

	return Handler(newmw(endpoint.Endpoint(hn)))
}

// newhandler returns a new handler
func newHandler(fn Handler, options ...HandlerOption) *handler {
	hn := &handler{
		encoder:      nil,
		decoder:      nil,
		errorEncoder: nil,
		errorhandler: nil,
		middlewares:  []Middleware{},
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
		kit_endpoint.Endpoint(
			Wrap(fn, hn.middlewares...),
		),
		kit_http.DecodeRequestFunc(hn.decoder),
		kit_http.EncodeResponseFunc(hn.encoder),
		hn.options...,
	)

	return hn
}
