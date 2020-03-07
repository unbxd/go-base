package http

import (
	"context"
	net_http "net/http"

	kit_http "github.com/go-kit/kit/transport/http"
)

// Decoder provides method to decode the body of the Request
type Decoder func(context.Context, *net_http.Request) (interface{}, error)

func newDefaultDecoder() Decoder {
	return func(ctx context.Context, r *net_http.Request) (interface{}, error) {
		return r, nil
	}
}

// NewDefaultDecoder returns default decoder for http
func NewDefaultDecoder() Decoder { return newDefaultDecoder() }

// NopRequestDecoder is no operation Decoder for the request
func NopRequestDecoder() Decoder {
	return Decoder(kit_http.NopRequestDecoder)
}

// NewGoKitDecoderHandlerOption sets a decoder of a type kit_http.DecodeRequestFunc
func NewGoKitDecoderHandlerOption(fn kit_http.DecodeRequestFunc) HandlerOption {
	return func(h *handler) {
		h.decoder = Decoder(fn)
	}
}
