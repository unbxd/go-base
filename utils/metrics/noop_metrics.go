package metrics

import (
	kit_metrics "github.com/go-kit/kit/metrics"
)

type noopMetricsCounter struct{}

func (nm *noopMetricsCounter) With(labelsValues ...string) kit_metrics.Counter {
	return &noopMetricsCounter{}
}
func (nm *noopMetricsCounter) Add(delta float64) {}

type noopMetricsHistogram struct{}

func (nm *noopMetricsHistogram) With(labelsValues ...string) kit_metrics.Histogram {
	return &noopMetricsHistogram{}
}
func (nm *noopMetricsHistogram) Observe(value float64) {}

type noopMetricsGauge struct{}

func (nm *noopMetricsGauge) With(labelsValues ...string) kit_metrics.Gauge {
	return &noopMetricsGauge{}
}

func (nm *noopMetricsGauge) Add(delta float64) {}

func (nm *noopMetricsGauge) Set(value float64) {}

type noopMetrics struct{}

func (nm noopMetrics) NewCounter(name string, sampleRate float64) Counter {
	return &noopMetricsCounter{}
}
func (nm noopMetrics) NewHistogram(name string, sampleRate float64) Histogram {
	return &noopMetricsHistogram{}
}
func (nm noopMetrics) NewGauge(name string) Gauge { return &noopMetricsGauge{} }

func NewNoopMetrics() Metrics { return &noopMetrics{} }
