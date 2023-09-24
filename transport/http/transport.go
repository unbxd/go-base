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

		muxOptions []MuxOption

		//server level filter, applicable for all handlers
		filters []Filter

		mux Mux

		logger log.Logger

		monitors []string
		metricer Metricser
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

	if tr.metricer != nil {
		if hn := tr.metricer.Handler(); hn != nil {
			tr.mux.Handler(net_http.MethodGet, "/metrics", hn)
		}
	}

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
func NewTransport(
	logger log.Logger,
	host, port string,
	options ...TransportOption,
) (*Transport, error) {
	transport := &Transport{
		Server:     &net_http.Server{Addr: host + ":" + port},
		options:    []HandlerOption{},
		mux:        nil,
		muxOptions: make([]MuxOption, 0),
		monitors:   []string{"/ping"},
		logger:     logger,
		filters: []Filter{
			CloserFilter(),
			PanicRecoveryFilter(logger),
			RequestIDFilter(),
			DecorateContextFilter(),
		},
	}

	for _, o := range options {
		o(transport)
	}

	if transport.mux == nil {
		transport.mux = NewDefaultMux(transport.muxOptions...)
	}

	transport.Handler = transport.mux

	if transport.filters != nil {
		transport.Handler = Chain(transport.mux, transport.filters...)
	}

	return transport, nil
}
