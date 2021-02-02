package http

import (
	"github.com/go-kit/kit/metrics"
	net_http "net/http"

	kitpr "github.com/go-kit/kit/metrics/prometheus"
	stdpr "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type defaultMetricer struct {
	namespace string
	fields    []string
}

func (dm *defaultMetricer) Counter(prefix, name string) metrics.Counter {
	return kitpr.NewCounterFrom(stdpr.CounterOpts{
		Namespace: dm.namespace,
		Subsystem: prefix,
		Name:      name,
		Help:      "namespace:" + dm.namespace + " subsystem:" + prefix + " name:" + name,
	}, dm.fields)
}

func (dm *defaultMetricer) Histogram(prefix, name string) metrics.Histogram {
	return kitpr.NewSummaryFrom(stdpr.SummaryOpts{
		Namespace: dm.namespace,
		Subsystem: prefix,
		Name:      name,
		Help:      "namespace:" + dm.namespace + " subsystem:" + prefix + " name:" + name,
	}, dm.fields)
}

func (dm *defaultMetricer) Handler() net_http.Handler {
	return promhttp.Handler()
}

// NewDefaultMetricer  returns metricer's default implementation
func NewDefaultMetricer(namespace string, fields []string) Metricser {
	return &defaultMetricer{namespace, fields}
}
