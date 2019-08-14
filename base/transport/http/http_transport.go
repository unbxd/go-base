package http

import (
	"context"
	net_http "net/http"
	"time"

	"github.com/go-kit/kit/endpoint"

	tmux "github.com/dimfeld/httptreemux"
	kit_http "github.com/go-kit/kit/transport/http"
	"github.com/uknth/go-base/base/log"
)

// TransportOption sets options parameters to Transport
type TransportOption func(*Transport)

// WithMux sets the multiplexer for transport
func WithMux(mux Mux) TransportOption {
	return func(tr *Transport) {
		tr.mux = mux
	}
}

// WithLogger sets custom logger for Transport
func WithLogger(logger log.Logger) TransportOption {
	return func(tr *Transport) {
		tr.logger = logger
	}
}

// WithFullDefaults sets default []kit_http.ServerOption, used
// by every request handler
func WithFullDefaults() TransportOption {
	return func(tr *Transport) {
		for _, opt := range []kit_http.ServerOption{
			NewRequestIDRequestFunc("Go-Base-Request-ID"),
			NewPopulateRequestContextRequestFunc(),
			NewCORSResponseFunc(),
			NewDefaultErrorEncoder(),
			NewTraceLoggerFinalizer(tr.logger),
		} {
			tr.options = append(tr.options, opt)
		}
	}
}

// WithOptionsOverride overrides the default []kit_http.ServerOption
// and replaces it with options provided
func WithOptionsOverride(options ...kit_http.ServerOption) TransportOption {
	return func(tr *Transport) { tr.options = options }
}

// WithMetricser supports adding metricer to Transport
func WithMetricser(metricer Metricser) TransportOption {
	return func(tr *Transport) { tr.metricer = metricer }
}

// WithOptionsAppend appends the provided kit_http.ServerOption(s)
// to existing ServerOption of transport
func WithOptionsAppend(options ...kit_http.ServerOption) TransportOption {
	return func(tr *Transport) { tr.options = append(tr.options, options...) }
}

// WithErrorEncoder lets us put a custom error encoder for the Transport
// If any Handler doesn't have an error encoder defined when throwing an error
// this error encoder is used
func WithErrorEncoder(errorEncoder kit_http.ErrorEncoder) TransportOption {
	return func(tr *Transport) { tr.errorEncoder = errorEncoder }
}

// WithTimeout sets the custom net_http.Server timeout for the Transport
func WithTimeout(idle, write, read time.Duration) TransportOption {
	return func(tr *Transport) {
		tr.IdleTimeout = idle
		tr.WriteTimeout = write
		tr.ReadTimeout = read
	}
}

// WithMonitors appends to a default list of monitor endpoint supported by Transport
func WithMonitors(monitors []string) TransportOption {
	return func(tr *Transport) {
		tr.monitors = append(tr.monitors, monitors...)
	}
}

// Transport is a wrapper around net_http.Server with sane defaults
// dimfeld/httptreemux is used as default multiplexer and Go-Kit's
// http transport is used as default request handler
type Transport struct {
	*net_http.Server

	// default ServerOptions
	options []kit_http.ServerOption

	mux          Mux
	logger       log.Logger
	monitors     []string
	metricer     Metricser
	errorEncoder kit_http.ErrorEncoder
}

func (tr *Transport) wrap(options []kit_http.ServerOption) []kit_http.ServerOption {
	return append(tr.options, options...)
}

// Get handles GET request
func (tr *Transport) Get(url string, fn HandlerFunc, options ...kit_http.ServerOption) {
	tr.mux.Handler(net_http.MethodGet, url, encapsulate(fn, tr.options, options))
}

// GetE handles GET request for custom definition of handler
func (tr *Transport) GetE(
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
func (tr *Transport) Put(url string, fn HandlerFunc, options ...kit_http.ServerOption) {
	tr.mux.Handler(net_http.MethodPut, url, encapsulate(fn, tr.options, options))
}

// PutE handles PUT request for custom definition of handler
func (tr *Transport) PutE(
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
func (tr *Transport) Post(url string, fn HandlerFunc, options ...kit_http.ServerOption) {
	tr.mux.Handler(net_http.MethodPost, url, encapsulate(fn, tr.options, options))
}

// PostE handles POST request for custom definition of handler
func (tr *Transport) PostE(
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
func (tr *Transport) Delete(url string, fn HandlerFunc, options ...kit_http.ServerOption) {
	tr.mux.Handler(net_http.MethodDelete, url, encapsulate(fn, tr.options, options))
}

// DeleteE handles DELETE request for custom definition of handler
func (tr *Transport) DeleteE(
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
func (tr *Transport) Patch(url string, fn HandlerFunc, options ...kit_http.ServerOption) {
	tr.mux.Handler(net_http.MethodPatch, url, encapsulate(fn, tr.options, options))
}

// PatchE handles PATCH request for custom definition of handler
func (tr *Transport) PatchE(
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
func (tr *Transport) Options(url string, fn HandlerFunc, options ...kit_http.ServerOption) {
	tr.mux.Handler(net_http.MethodOptions, url, encapsulate(fn, tr.options, options))
}

// OptionsE handles OPTIONS request for custom definition of handler
func (tr *Transport) OptionsE(
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
func (tr *Transport) Head(url string, fn HandlerFunc, options ...kit_http.ServerOption) {
	tr.mux.Handler(net_http.MethodHead, url, encapsulate(fn, tr.options, options))
}

// HeadE handles HEAD request for custom definition of handler
func (tr *Transport) HeadE(
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
func (tr *Transport) Trace(url string, fn HandlerFunc, options ...kit_http.ServerOption) {
	tr.mux.Handler(net_http.MethodTrace, url, encapsulate(fn, tr.options, options))
}

// TraceE handles TRACE request for custom definition of handler
func (tr *Transport) TraceE(
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
func (tr *Transport) Handle(method, url string, fn HandlerFunc, options ...kit_http.ServerOption) {
	tr.mux.Handler(method, url, encapsulate(fn, tr.options, options))
}

// Open starts the Transport
func (tr *Transport) Open() error {
	tr.Handler = tr.mux

	if hn := tr.metricer.Handler(); hn != nil {
		tr.mux.Handler(net_http.MethodGet, "/metrics", hn)
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
) *Transport {
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

	return transport
}
