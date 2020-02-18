package http

import (
	"context"
	net_http "net/http"
	"time"

	"github.com/go-kit/kit/endpoint"

	tmux "github.com/dimfeld/httptreemux"
	kit_http "github.com/go-kit/kit/transport/http"
	"github.com/unbxd/go-base/base/log"
)

type (
	// TransportOption sets options parameters to Transport
	TransportOption func(*Transport)

	// HandlerOption is set of options supported by the Transport
	// per Handler
	HandlerOption kit_http.ServerOption

	// Transport is a wrapper around net_http.Server with sane defaults
	// dimfeld/httptreemux is used as default multiplexer and Go-Kit's
	// http transport is used as default request handler
	Transport struct {
		*net_http.Server

		// default ServerOptions
		options []kit_http.ServerOption

		mux          Mux
		logger       log.Logger
		monitors     []string
		metricer     Metricser
		errorEncoder kit_http.ErrorEncoder
	}
)

func (tr *Transport) wrap(options []kit_http.ServerOption) []kit_http.ServerOption {
	return append(tr.options, options...)
}

// Get handles GET request
func (tr *Transport) Get(url string, fn HandlerFunc, options ...HandlerOption) {
	tr.mux.Handler(net_http.MethodGet, url, encapsulate(fn, tr.options, options))
}

// KitGET handles GET request for custom definition of handler.
// It exposes custom APIs defined as part of Go-Kit.
func (tr *Transport) KitGET(
	url string,
	decoder kit_http.DecodeRequestFunc,
	endpoint endpoint.Endpoint,
	encoder kit_http.EncodeResponseFunc,
	options ...kit_http.ServerOption,
) {
	tr.mux.Handler(
		net_http.MethodGet,
		url,
		kit_http.NewServer(
			endpoint, decoder, encoder, tr.wrap(options)...,
		),
	)
}

// Put handles PUT request
func (tr *Transport) Put(url string, fn HandlerFunc, options ...HandlerOption) {
	tr.mux.Handler(net_http.MethodPut, url, encapsulate(fn, tr.options, options))
}

// KitPUT handles PUT request for custom definition of handler
// It exposes custom APIs defined as part of Go-Kit.
func (tr *Transport) KitPUT(
	url string,
	decoder kit_http.DecodeRequestFunc,
	endpoint endpoint.Endpoint,
	encoder kit_http.EncodeResponseFunc,
	options ...kit_http.ServerOption,
) {
	tr.mux.Handler(
		net_http.MethodPut,
		url,
		kit_http.NewServer(
			endpoint, decoder, encoder, tr.wrap(options)...,
		),
	)
}

// Post handles POST request
func (tr *Transport) Post(url string, fn HandlerFunc, options ...HandlerOption) {
	tr.mux.Handler(net_http.MethodPost, url, encapsulate(fn, tr.options, options))
}

// KitPOST handles POST request for custom definition of handler
// It exposes custom APIs defined as part of Go-Kit.
func (tr *Transport) KitPOST(
	url string,
	decoder kit_http.DecodeRequestFunc,
	endpoint endpoint.Endpoint,
	encoder kit_http.EncodeResponseFunc,
	options ...kit_http.ServerOption,
) {
	tr.mux.Handler(
		net_http.MethodPost,
		url,
		kit_http.NewServer(
			endpoint, decoder, encoder, tr.wrap(options)...,
		),
	)
}

// Delete handles DELETE request
func (tr *Transport) Delete(url string, fn HandlerFunc, options ...HandlerOption) {
	tr.mux.Handler(net_http.MethodDelete, url, encapsulate(fn, tr.options, options))
}

// KitDELETE handles DELETE request for custom definition of handler
// It exposes custom APIs defined as part of Go-Kit.
func (tr *Transport) KitDELETE(
	url string,
	decoder kit_http.DecodeRequestFunc,
	endpoint endpoint.Endpoint,
	encoder kit_http.EncodeResponseFunc,
	options ...kit_http.ServerOption,
) {
	tr.mux.Handler(
		net_http.MethodDelete,
		url,
		kit_http.NewServer(
			endpoint, decoder, encoder, tr.wrap(options)...,
		),
	)
}

// Patch handles PATCH request
func (tr *Transport) Patch(url string, fn HandlerFunc, options ...HandlerOption) {
	tr.mux.Handler(net_http.MethodPatch, url, encapsulate(fn, tr.options, options))
}

// KitPATCH handles PATCH request for custom definition of handler
// It exposes custom APIs defined as part of Go-Kit.
func (tr *Transport) KitPATCH(
	url string,
	decoder kit_http.DecodeRequestFunc,
	endpoint endpoint.Endpoint,
	encoder kit_http.EncodeResponseFunc,
	options ...kit_http.ServerOption,
) {
	tr.mux.Handler(
		net_http.MethodPatch,
		url,
		kit_http.NewServer(
			endpoint, decoder, encoder, tr.wrap(options)...,
		),
	)
}

// Options handles OPTIONS request
func (tr *Transport) Options(url string, fn HandlerFunc, options ...HandlerOption) {
	tr.mux.Handler(net_http.MethodOptions, url, encapsulate(fn, tr.options, options))
}

// KitOPTION handles OPTIONS request for custom definition of handler
// It exposes custom APIs defined as part of Go-Kit.
func (tr *Transport) KitOPTION(
	url string,
	decoder kit_http.DecodeRequestFunc,
	endpoint endpoint.Endpoint,
	encoder kit_http.EncodeResponseFunc,
	options ...kit_http.ServerOption,
) {
	tr.mux.Handler(
		net_http.MethodOptions,
		url,
		kit_http.NewServer(
			endpoint, decoder, encoder, tr.wrap(options)...,
		),
	)
}

// Head handles HEAD request
func (tr *Transport) Head(url string, fn HandlerFunc, options ...HandlerOption) {
	tr.mux.Handler(net_http.MethodHead, url, encapsulate(fn, tr.options, options))
}

// KitHEAD handles HEAD request for custom definition of handler
// It exposes custom APIs defined as part of Go-Kit.
func (tr *Transport) KitHEAD(
	url string,
	decoder kit_http.DecodeRequestFunc,
	endpoint endpoint.Endpoint,
	encoder kit_http.EncodeResponseFunc,
	options ...kit_http.ServerOption,
) {
	tr.mux.Handler(
		net_http.MethodHead,
		url,
		kit_http.NewServer(
			endpoint, decoder, encoder, tr.wrap(options)...,
		),
	)
}

// Trace handles TRACE request
func (tr *Transport) Trace(url string, fn HandlerFunc, options ...HandlerOption) {
	tr.mux.Handler(net_http.MethodTrace, url, encapsulate(fn, tr.options, options))
}

// KitTRACE handles TRACE request for custom definition of handler
// It exposes custom APIs defined as part of Go-Kit.
func (tr *Transport) KitTRACE(
	url string,
	decoder kit_http.DecodeRequestFunc,
	endpoint endpoint.Endpoint,
	encoder kit_http.EncodeResponseFunc,
	options ...kit_http.ServerOption,
) {
	tr.mux.Handler(
		net_http.MethodTrace,
		url,
		kit_http.NewServer(
			endpoint, decoder, encoder, tr.wrap(options)...,
		),
	)
}

// Handle is generic method to allow custom bindings of URL with a method and it's handler
func (tr *Transport) Handle(method, url string, fn HandlerFunc, options ...HandlerOption) {
	tr.mux.Handler(method, url, encapsulate(fn, tr.options, options))
}

// ServerOptions returns the default server options associated with Transport
func (tr *Transport) ServerOptions() []kit_http.ServerOption { return tr.options }

// Open starts the Transport
func (tr *Transport) Open() error {
	tr.Handler = tr.mux

	if tr.metricer != nil {
		if hn := tr.metricer.Handler(); hn != nil {
			tr.mux.Handler(net_http.MethodGet, "/metrics", hn)
		}
	}

	for _, mon := range tr.monitors {
		tr.mux.Handler(net_http.MethodGet, mon, net_http.HandlerFunc(
			func(rw net_http.ResponseWriter, req *net_http.Request) {
				rw.WriteHeader(net_http.StatusOK)
				rw.Write([]byte("alive"))
			},
		))
	}

	return tr.ListenAndServe()
}

// Close shuts down Transport
func (tr *Transport) Close() error {
	ctx, cancel := context.WithTimeout(
		context.Background(), 100*time.Second,
	)

	defer cancel()
	return tr.Shutdown(ctx)
}

// NewTransport returns a new transport
func NewTransport(
	host, port string,
	options ...TransportOption,
) (*Transport, error) {
	transport := &Transport{
		Server:       &net_http.Server{Addr: host + ":" + port},
		options:      []kit_http.ServerOption{},
		mux:          tmux.NewContextMux(),
		monitors:     []string{"/ping"},
		errorEncoder: kit_http.DefaultErrorEncoder,
	}

	for _, o := range options {
		o(transport)
	}

	transport.options = append(
		transport.options,
		kit_http.ServerErrorEncoder(transport.errorEncoder),
	)

	return transport, nil
}
