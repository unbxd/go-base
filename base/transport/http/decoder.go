package http

import (
	"context"
	net_http "net/http"
)

// Decoder provides method to decode the body of the Request
type Decoder func(context.Context, *net_http.Request) (interface{}, error)

func newDefaultDecoder() Decoder {
	return func(ctx context.Context, r *net_http.Request) (interface{}, error) {
		return r, nil
	}
}
