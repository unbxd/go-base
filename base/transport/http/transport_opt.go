package http

import (
	"time"

	kit_http "github.com/go-kit/kit/transport/http"
	"github.com/vtomar01/go-base/base/log"
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

// WithFullDefaults sets default ServerOption, used
// by every request handler
// It sets following filters for the request
// 	- RequestID
//	- CORS
// 	- DefaultErrorHandler
//	- DefaultTranceLogger (using transport.Logger)
func WithFullDefaults() TransportOption {
	return func(tr *Transport) {
		for _, opt := range []HandlerOption{
			NewRequestIDHandlerOption("Go-Base-Request-ID"),
			NewCORSHandlerOption(),
			NewErrorEncoderHandlerOptions(kit_http.DefaultErrorEncoder),
			NewTraceLoggerFinalizerHandlerOption(tr.logger),
		} {
			tr.options = append(tr.options, opt)
		}
	}
}

// WithHandlerOption overrides the default HandlerOption for the transport
func WithHandlerOption(options ...HandlerOption) TransportOption {
	return func(tr *Transport) {
		tr.options = append(tr.options, options...)
	}
}

// WithMetricser supports adding metricer to Transport
func WithMetricser(metricer Metricser) TransportOption {
	return func(tr *Transport) { tr.metricer = metricer }
}

// WithTransportErrorEncoder lets us put a custom error encoder for the Transport
// applicable at Transport level. There is a provision to do this per handler
// using NewErrorEncoder, however if a handler doesn't have an error encoder
// this will be used as default
// If any Handler doesn't have an error encoder defined when throwing an error
// this error encoder is used
func WithTransportErrorEncoder(fn ErrorEncoder) TransportOption {
	return func(tr *Transport) {
		tr.options = append(
			tr.options,
			NewErrorEncoderHandlerOptions(fn),
		)
	}
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
