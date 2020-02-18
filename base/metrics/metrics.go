package metrics

import (
	kit_metrics "github.com/go-kit/kit/metrics"
)

type (
	// Counter extends kit_metrics.Counter
	Counter interface {
		kit_metrics.Counter
	}

	// Gauge extends kit_metrics.Gauge
	Gauge interface {
		kit_metrics.Gauge
	}

	// Histogram extends kit_metrics.Histogram
	Histogram interface {
		kit_metrics.Histogram
	}

	// Metrics is wrapper for supported metrics interface
	Metrics interface {
		NewCounter(name string, sampleRate float64) Counter

		NewGauge(name string) Gauge

		NewHistogram(name string, sampleRate float64) Histogram
	}
)
