package http

import (
	"github.com/go-kit/kit/metrics"
	net_http "net/http"
)

// Mux defines the standard Multiplexer for http Request
type Mux interface {
	// ServeHTTP
	net_http.Handler

	// Default Handler Method is all that is needed
	Handler(method, url string, fn net_http.Handler)
}

// Metricser is wrapper for supported metrics agents
type Metricser interface {
	Counter(prefix, name string) metrics.Counter
	Histogram(prefix, name string) metrics.Histogram
	Handler() net_http.Handler
}
