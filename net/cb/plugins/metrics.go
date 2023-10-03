package plugins

import (
	gbmetrics "github.com/unbxd/go-base/metrics"
	"github.com/unbxd/hystrix-go/hystrix/metric"
)

// Metric Names
const (
	CircuitOpen       = "cb.circuitOpen"
	Attempts          = "cb.attempts"
	Errors            = "cb.errors"
	Successes         = "cb.successes"
	Failures          = "cb.failures"
	Rejects           = "cb.rejects"
	ShortCircuits     = "cb.shortCircuits"
	Timeouts          = "cb.timeouts"
	FallbackSuccesses = "cb.fallbackSuccesses"
	FallbackFailures  = "cb.fallbackFailures"
	TotalDuration     = "cb.totalDuration"
	RunDuration       = "cb.runDuration"
)

type metricsCollector struct {
	lvls []string

	attemptsCounter          gbmetrics.Counter
	errorsCounter            gbmetrics.Counter
	successCounter           gbmetrics.Counter
	failuresCounter          gbmetrics.Counter
	rejectsCounter           gbmetrics.Counter
	shortcircuitsCounter     gbmetrics.Counter
	timeoutsCounter          gbmetrics.Counter
	fallbackSuccessesCounter gbmetrics.Counter
	fallbackFailuresCounter  gbmetrics.Counter

	circuitOpenGauge gbmetrics.Gauge

	totalDurationHistogram gbmetrics.Histogram
	rundurationHistogram   gbmetrics.Histogram
}

func (mc *metricsCollector) Update(r metric.Result) {
	if r.Attempts > 0 {
		mc.attemptsCounter.With(mc.lvls...).Add(r.Attempts)
	}
	if r.Errors > 0 {
		mc.errorsCounter.With(mc.lvls...).Add(r.Errors)
	}
	if r.Successes > 0 {
		mc.circuitOpenGauge.With(mc.lvls...).Set(0)
		mc.successCounter.With(mc.lvls...).Add(r.Successes)
	}
	if r.Failures > 0 {
		mc.failuresCounter.With(mc.lvls...).Add(r.Failures)
	}
	if r.Rejects > 0 {
		mc.rejectsCounter.With(mc.lvls...).Add(r.Rejects)
	}
	if r.ShortCircuits > 0 {
		mc.circuitOpenGauge.With(mc.lvls...).Add(1)
		mc.shortcircuitsCounter.With(mc.lvls...).Add(r.ShortCircuits)
	}
	if r.Timeouts > 0 {
		mc.timeoutsCounter.With(mc.lvls...).Add(r.Timeouts)
	}
	if r.FallbackSuccesses > 0 {
		mc.fallbackSuccessesCounter.With(mc.lvls...).Add(r.FallbackSuccesses)
	}
	if r.FallbackFailures > 0 {
		mc.fallbackFailuresCounter.With(mc.lvls...).Add(r.FallbackFailures)
	}

	mc.totalDurationHistogram.With(mc.lvls...).Observe(float64(r.TotalDuration.Milliseconds()))
	mc.rundurationHistogram.With(mc.lvls...).Observe(float64(r.RunDuration.Milliseconds()))
}

func (mc *metricsCollector) Reset() {}

// NewMetricsCollector returns a wrapped collector for go-base/utils/metrics.Metrics
func NewMetricsCollector(metrics gbmetrics.Provider) func(string) metric.Collector {
	collector := &metricsCollector{
		attemptsCounter:          metrics.NewCounter(Attempts, 1.0),
		errorsCounter:            metrics.NewCounter(Errors, 1.0),
		successCounter:           metrics.NewCounter(Successes, 1.0),
		failuresCounter:          metrics.NewCounter(Failures, 1.0),
		rejectsCounter:           metrics.NewCounter(Rejects, 1.0),
		shortcircuitsCounter:     metrics.NewCounter(ShortCircuits, 1.0),
		timeoutsCounter:          metrics.NewCounter(Timeouts, 1.0),
		fallbackSuccessesCounter: metrics.NewCounter(FallbackSuccesses, 1.0),
		fallbackFailuresCounter:  metrics.NewCounter(FallbackFailures, 1.0),
		circuitOpenGauge:         metrics.NewGauge(CircuitOpen),
		totalDurationHistogram:   metrics.NewHistogram(TotalDuration, 1.0),
		rundurationHistogram:     metrics.NewHistogram(RunDuration, 1.0),
	}

	return func(name string) metric.Collector {
		collector.lvls = append(collector.lvls, []string{"cb", name}...)
		return collector
	}
}
