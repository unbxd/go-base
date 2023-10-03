package metrics

import (
	net_http "net/http"

	kit_metrics "github.com/go-kit/kit/metrics"
)

type (
	Counter   interface{ kit_metrics.Counter }
	Gauge     interface{ kit_metrics.Gauge }
	Histogram interface{ kit_metrics.Histogram }

	// Handler interface exposes metrics which support handler
	Handler interface{ Handler() net_http.Handler }

	// Provider standarizes the metrics interface used by the applications
	Provider interface {
		NewCounter(name string, sampleRate float64) Counter
		NewHistogram(name string, sampleRate float64) Histogram
		NewGauge(name string) Gauge
	}
)
