package http

import (
	"context"
	net_http "net/http"
	"time"

	"github.com/unbxd/go-base/log"
)

type (
	// TransportOption sets options parameters to Transport
	TransportOption func(*Transport)

	// Transport is a wrapper around net_http.Server with sane defaults
	// dimfeld/httptreemux is used as default multiplexer and Go-Kit's
	// http transport is used as default request handler
	Transport struct {
		*net_http.Server

		// default HandlerOption
		options []HandlerOption

		//server level filter, applicable for all handlers
		filters []Filter

		mux Mux

		logger log.Logger

		monitors []string
	}
)

// Mux returns the default multiplexer
func (tr *Transport) Mux() Mux { return tr.mux }

// TransportWithFilter sets up filters for the Transport
func TransportWithFilter(f Filter) TransportOption {
	return func(tr *Transport) {
		tr.filters = append(
			tr.filters, f,
		)
	}
}

// Open starts the Transport
func (tr *Transport) Open() error {

	for _, mon := range tr.monitors {
		tr.mux.Handler(net_http.MethodGet, mon, net_http.HandlerFunc(
			func(rw net_http.ResponseWriter, req *net_http.Request) {
				rw.WriteHeader(net_http.StatusOK)
				rw.Write([]byte("alive")) //nolint:errcheck
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
// TODO: use Transport Config to build the Transport, configurations shouldn't be like this
//
// Deprecated: use the new config interface to create Transport.
// This one tries to do the best with sane defaults, but the configuration
// is way more streamlined in the new intializer `NewHttpTansport`.
func NewTransport(
	host, port string,
	options ...TransportOption,
) (*Transport, error) {
	logger, _ := log.NewZapLogger()

	transport := &Transport{
		Server:   &net_http.Server{Addr: host + ":" + port},
		options:  []HandlerOption{},
		mux:      nil,
		monitors: []string{"/ping"},
		logger:   logger,
		filters:  []Filter{},
	}

	for _, o := range options {
		o(transport)
	}

	if transport.mux == nil {
		transport.mux = NewDefaultMux() // just defaults with chi
	}

	transport.Handler = transport.mux

	filters := []Filter{
		CloserFilter(),                // Simple Closer, closes the request in defer
		WrappedResponseWriterFilter(), // casts to WrappedResposeWriter, gives bunch of util functions
		PanicRecoveryFilter(transport.logger, WithStack(1024*8, false)), // Handles Panic
		RequestIDFilter(),       // Handles RequestID generation
		DecorateContextFilter(), // Decorates the context.Context of request
		OpenTelemetryFilterForDefaultMux([]string{}, make(map[string]string)), // Open telemetry
		TraceLoggingFilter(transport.logger),                                  // Trace Logging
	}

	if transport.filters != nil {
		filters = append(filters, transport.filters...)
	}

	transport.Handler = Chain(transport.mux, filters...)
	return transport, nil
}
