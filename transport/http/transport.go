package http

import (
	"context"
	http "net/http"
	"time"

	"github.com/unbxd/go-base/v2/log"
)

type (
	// TransportOption sets options parameters to Transport
	TransportOption func(*Transport)

	// Transport is a wrapper around http.Server with sane defaults
	// dimfeld/httptreemux is used as default multiplexer and Go-Kit's
	// http transport is used as default request handler
	Transport struct {
		*http.Server

		name string

		logger log.Logger
		muxer  Muxer

		options []HandlerOption
	}
)

// Mux returns the default multiplexer
func (tr *Transport) Mux() Muxer { return tr.muxer }

// Open starts the Transport
func (tr *Transport) Open() error {
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
//
// Deprecated: use the new config interface to create Transport.
// This one tries to do the best with sane defaults, but the configuration
// is way more streamlined in the new intializer `NewHttpTansport`.
func NewTransport(
	host, port string,
	options ...TransportOption,
) (*Transport, error) {
	return NewHTTPTransport(
		"gobi",
		WithCustomHostPort(host, port),
		WithNoMetrics(), // default metrics are turned off, use your own
		WithDefaultTransportOptions(options...),
	)
}
