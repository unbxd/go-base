package http

import (
	"time"

	kit_http "github.com/go-kit/kit/transport/http"
	"github.com/unbxd/go-base/base/log"
)

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
