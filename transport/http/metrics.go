package http

import (
	net_http "net/http"

	"github.com/go-kit/kit/metrics"
)

// Metricser is wrapper for supported metrics agents
type Metricser interface {
	Counter(prefix, name string) metrics.Counter
	Histogram(prefix, name string) metrics.Histogram
	Handler() net_http.Handler
}
