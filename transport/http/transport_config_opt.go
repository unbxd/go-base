package http

import (
	"net/http"
	"time"

	"github.com/unbxd/go-base/v2/log"
	"github.com/unbxd/go-base/v2/metrics"
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

type OpenTelemetryProvider interface {
	metric.MeterProvider
	trace.TracerProvider
}

type OpenTelemetryRequestFilter otelhttp.Filter

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
	enabled bool,
	provider OpenTelemetryProvider,
	tags []KeyValue,
	filters ...OpenTelemetryRequestFilter,
) TransportConfigOption {
	return func(c *config) (err error) {
		if enabled {
			ff := []OpenTelemetryRequestFilter{
				OpenTelemetryRequestFilter(filterMonitors(c.heartbeats)),
			}

			ff = append(ff, filters...)

			c.ffs = append(c.ffs, OpenTelemetryFilter(
				c.name,
				provider,
				tags,
				ff...,
			))

		}
		return
	}
}

// WithCustomMetrics lets you use `metrics.Counter`, `metrics.Histogram` & `metrics.Gauge` interfaces
// instead of relying on some third party interfaces. Not passing custom formatter will result in
// default formatter being used.
func WithCustomMetrics(
	enabled bool,
	provider metrics.Provider,
	formatter MetricsNameFormatter,
) TransportConfigOption {
	return func(c *config) (err error) {
		if enabled {
			c.ffs = append(c.ffs, CustomMetricsFilter(
				c.name,
				provider,
				formatter,
			))
		}
		return
	}
}

// WithTransportOption can be used to set other overridable Transport Options
func WithTransportOption(options ...TransportOption) TransportConfigOption {
	return func(c *config) (err error) {
		c.transportOptions = append(c.transportOptions, options...)
		return
	}
}

func WithHandlerOptionForTransport(options ...HandlerOption) TransportConfigOption {
	return func(c *config) (err error) {
		for _, o := range options {
			c.transportOptions = append(c.transportOptions, WithHandlerOption(o))
		}
		return nil
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

func WithTraceLogging(fieldsGens ...TraceLogFieldsGen) TransportConfigOption {
	return func(c *config) error {
		c.ffs = append(c.ffs, TraceLoggingFilter(c.logger, fieldsGens...))
		return nil
	}
}
