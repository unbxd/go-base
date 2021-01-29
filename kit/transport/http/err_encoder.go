package http

import (
	"context"
	net_http "net/http"

	kit_http "github.com/go-kit/kit/transport/http"
)

// ErrorEncoder defines the Encoder that handles error
type ErrorEncoder func(context.Context, error, net_http.ResponseWriter)

// WithErrorEncoder sets error handler for the error generated
// by the request
func WithErrorEncoder(fn ErrorEncoder) TransportOption {
	return func(tr *Transport) {
		tr.options = append(
			tr.options, NewErrorEncoderHandlerOptions(fn),
		)
	}
}

// NewErrorEncoderHandlerOptions provides a handler specific
// error encoder
func NewErrorEncoderHandlerOptions(fn ErrorEncoder) HandlerOption {
	return func(h *handler) {
		h.errorEncoder = fn
		h.options = append(
			h.options,
			kit_http.ServerErrorEncoder(
				kit_http.ErrorEncoder(fn),
			),
		)
	}
}

// NewGoKitErrorEncoderHandlerOption provides an option to set
// error encoder for handler or transport based on Go-Kit's ErrorEncoder
func NewGoKitErrorEncoderHandlerOption(fn kit_http.ErrorEncoder) HandlerOption {
	return func(h *handler) {
		h.errorEncoder = ErrorEncoder(fn)
	}
}
