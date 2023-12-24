package http

import (
	"time"
)

// WithMuxer sets the multiplexer for transport
func WithMuxer(mux Muxer) TransportOption {
	return func(tr *Transport) {
		tr.muxer = mux
	}
}

// WithFullDefaults sets default ServerOption, used
// by every request handler
// It sets following filters for the request
//   - RequestID
//   - CORS
//   - DefaultErrorHandler
//   - DefaultTranceLogger (using transport.Logger)
//
// Deprecated: use `WithProductionDefaults` for Production Environments, `WithDevDefaults` for Dev Env
func WithFullDefaults() TransportOption {
	return func(tr *Transport) {
		tr.handlerOptions = append(tr.handlerOptions, []HandlerOption{}...)
	}
}

// WithHandlerOption overrides the default HandlerOption for the transport
func WithHandlerOption(options ...HandlerOption) TransportOption {
	return func(tr *Transport) {
		tr.handlerOptions = append(tr.handlerOptions, options...)
	}
}

// WithTransportErrorEncoder lets us put a custom error encoder for the Transport
// applicable at Transport level. There is a provision to do this per handler
// using NewErrorEncoder, however if a handler doesn't have an error encoder
// this will be used as default
// If any Handler doesn't have an error encoder defined when throwing an error
// this error encoder is used
// Deprecated: default error handler is not overridable, defaultErrorHandler will
// become default in next release
func WithTransportErrorEncoder(fn ErrorEncoder) TransportOption {
	return func(tr *Transport) {
		tr.handlerOptions = append(
			tr.handlerOptions,
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
