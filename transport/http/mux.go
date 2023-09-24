package http

import (
	net_http "net/http"

	"github.com/go-chi/chi/v5"
)

// Mux defines the standard Multiplexer for http Request
type Mux interface {
	// ServeHTTP
	net_http.Handler

	// Default Handler Method is all that is needed
	Handler(method, url string, fn net_http.Handler)
}

type MuxOption func(*muxer)

type muxer struct {
	*chi.Mux
}

func (mx *muxer) Handler(method, url string, fn net_http.Handler) {
	mx.Method(method, url, fn)
}

func NewDefaultMux(opts ...MuxOption) Mux {
	mx := &muxer{chi.NewMux()}

	for _, o := range opts {
		o(mx)
	}

	return mx
}
