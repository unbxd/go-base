package http

import (
	"net/http"
	"time"

	"github.com/unbxd/go-base/log"
	"github.com/unbxd/go-base/metrics"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

func WithVersion(version string) TransportConfigOption {
	return func(c *config) (err error) {
		c.version = version
		return
	}
}

func WithCustomHostPort(host, port string) TransportConfigOption {
	return func(c *config) (err error) {
		c.host = host
		c.port = port
		return
	}
}

// WithHeartbeats sets list of path which returns a simple 200 & ping response
func WithHeartbeats(heartbeats ...string) TransportConfigOption {
	return func(c *config) (err error) {
		c.heartbeats = append(c.heartbeats, heartbeats...)
		return
	}
}

func WithCustomTimeouts(idle, read, write time.Duration) TransportConfigOption {
	return func(c *config) (err error) {
		c.idleTimeout = idle
		c.writeTimeout = write
		c.readTimeout = read
		return
	}
}

// WithCustomLogger uses the logger passed as an argument
func WithCustomLogger(logger log.Logger) TransportConfigOption {
	return func(c *config) (err error) {
		c.logger = logger
		return
	}
}

// WithLogger creates a new logger and associates it with Transport
func WithLogger(level string) TransportConfigOption {
	return func(c *config) (err error) {
		lgr, err := log.NewZeroLogger(log.ZeroLoggerWithLevel(level))
		if err != nil {
			return err
		}

		c.logger = lgr
		return
	}
}

// WithNoTraceLogging disables Trace-Logging of the server
func WithNoTraceLogging() TransportConfigOption {
	return func(c *config) (err error) {
		c.traceLogging = false
		return
	}
}

// WithNoLogging disables logging completely
func WithNoLogging() TransportConfigOption {
	return func(c *config) (err error) {
		c.logging = false
		return
	}
}

// WithNoMetrics disables all metrics emitted by the Transport
func WithNoMetrics() TransportConfigOption {
	return func(c *config) (err error) {
		c.metrics = false
		return
	}
}

type OpenTelemetryProvider interface {
	metric.MeterProvider
	trace.TracerProvider
}

// filters all monitor APIs
func filterMonitors(monitors []string) otelhttp.Filter {
	return func(r *http.Request) bool {
		path := r.URL.Path

		for _, m := range monitors {
			if path == m {
				return false
			}
		}

		return true
	}
}

func WithOpenTelemetryMetrics(
	provider OpenTelemetryProvider,
	tags []KeyValue,
	filters ...otelhttp.Filter,
) TransportConfigOption {
	return func(c *config) (err error) {
		if c.metrics {
			ff := []otelhttp.Filter{
				filterMonitors(c.heartbeats),
			}

			ff = append(ff, filters...)

			c.ffs = append(c.ffs, OpenTelemetryFilterForDefaultMux(
				c.name,
				tags,
				provider,
				provider,
				ff...,
			))

		}
		return
	}
}

// WithCustomMetrics lets you use `metrics.Counter`, `metrics.Histogram` & `metrics.Gauge` interfaces
// instead of relying on some third party interfaces. Not passing custom formatter will result in
// default formatter being used.
func WithCustomMetrics(provider metrics.Provider, formatter MetricsNameFormatter, tags ...KeyValue) TransportConfigOption {
	return func(c *config) (err error) {
		if c.metrics {
			c.ffs = append(c.ffs, CustomMetricsForDefaultMuxFilter(
				c.name,
				provider,
				formatter,
				tags...,
			))
		}
		return
	}
}

// WithDefaultTransportOptions can be used to set other overridable Transport Options
func WithDefaultTransportOptions(options ...TransportOption) TransportConfigOption {
	return func(c *config) (err error) {
		c.transportOptions = append(c.transportOptions, options...)
		return
	}
}

// WithCustomHeadersForRequestID lets you configure what all headers should have
// request ID. It is useful for cases where internal tracking is based on
func WithCustomHeadersForRequestID(formatter RequestIDFormatter, headers ...string) TransportConfigOption {
	return func(c *config) (err error) {
		c.ffs = append(
			c.ffs,
			CustomRequestIDFilter(
				formatter, headers...,
			),
		)
		return
	}
}

// WithFilters allows to add custom Filter to the Transport
func WithFilters(filters ...Filter) TransportConfigOption {
	return func(c *config) (err error) {
		c.ffs = append(c.ffs, filters...)
		return
	}
}

func WithCustomPanicFormatter(formatter PanicFormatter) TransportConfigOption {
	return func(c *config) (err error) {
		c.panicFormatter = formatter
		return
	}
}

func WithDefaultPanicFormatter(panicFormatterType PanicFormatterType) TransportConfigOption {
	return func(c *config) (err error) {
		switch panicFormatterType {
		case HTMLPanicFormatter:
			c.panicFormatter = &htmlPanicFormatter{}
		case TextPanicFormatter:
			c.panicFormatter = &textPanicFormatter{}
		default:
			c.panicFormatter = &textPanicFormatter{}
		}
		return
	}
}
