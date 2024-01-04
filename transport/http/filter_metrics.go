package http

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/unbxd/go-base/v2/metrics"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type SpanNameFormatter func(operation string, r *http.Request) string

func DefaultSpanNameFormatter(operation string, r *http.Request) string {
	// we will only get this if chi is the router
	var (
		sb  strings.Builder
		rpt = chi.RouteContext(r.Context()).RoutePattern()
	)

	if operation != "" {
		sb.WriteString(operation)
		sb.WriteRune(' ')
	}

	sb.WriteString(r.Method)
	sb.WriteRune(' ')
	sb.WriteString(rpt)

	return sb.String()
}

// OpenTelemetryFilter uses OpenTelemetry to publish events
// There are multiple providers for OpenTelemetry that can be used
// A simple example of using this filter is by just setting this up in the
// filter chain and in the application, set the provider
// Example using Datadog
// gopkg.in/DataDog/dd-trace-go.v1/ddtrace/opentelemetry
//
//	import (
//		"go.opentelemetry.io/otel"
//		ddotel "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/opentelemetry"
//	)
//
//	func main() {
//		provider := ddotel.NewTracerProvider()
//		defer provider.Shutdown()
//		otel.SetTracerProvider(provider)
//	}
func OpenTelemetryFilter(
	namespace string,
	provider OpenTelemetryProvider,
	tags []KeyValue,
	filters ...OpenTelemetryRequestFilter,
) Filter {
	formatter := DefaultSpanNameFormatter
	attribs := make([]attribute.KeyValue, 0)

	for _, kv := range tags {
		attribs = append(attribs, attribute.String(kv.Key, kv.Value))
	}

	options := []otelhttp.Option{}

	for _, fn := range filters {
		options = append(options, otelhttp.WithFilter(otelhttp.Filter(fn)))
	}

	options = append(options, []otelhttp.Option{
		otelhttp.WithSpanNameFormatter(formatter),
		otelhttp.WithSpanOptions(
			trace.WithNewRoot(),
			trace.WithAttributes(attribs...),
		),
		otelhttp.WithMeterProvider(provider),
		otelhttp.WithTracerProvider(provider),
	}...)

	// this is slightly in-efficient that we are double wrapping
	// http.ResponseWriter.
	// In this middleware, there is a wrapping of 'respWriterWrapper'
	// to extract the status code, which we do too in case of
	// `WrapResponseWriter`
	// This in itself shouldn't cause any issue because ResponseWriter
	// is an interface and we are wrapping it twice, just that
	// it introduces a bit of overhead of computation
	return otelhttp.NewMiddleware(
		namespace+"::http-serve",
		options...,
	)
}

type MetricsNameFormatter func(namespace string, r *http.Request) string

type MetricsTagGenerator func(rw WrapResponseWriter, req *http.Request) []KeyValue

func DefaultMetricsNameFormatter(namespace string, r *http.Request) string {
	rcx := chi.RouteContext(r.Context())
	if rcx == nil {
		return namespace + ".not-chi"
	}

	var sb strings.Builder

	rpt := rcx.RoutePattern()

	sb.WriteRune('.')
	sb.WriteString(strings.ReplaceAll(rpt, "/", "_"))

	return sb.String()
}

func CustomMetricsFilter(
	namespace string,
	provider metrics.Provider,
	formatter MetricsNameFormatter,
	tagsGenerators ...MetricsTagGenerator,
) Filter {
	var (
		// counters   = make(map[string]metrics.Counter)
		histograms = make(map[string]metrics.Histogram)
	)

	if formatter == nil {
		formatter = DefaultMetricsNameFormatter
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			defer func() {
				var (
					label = formatter(namespace, r)
					tags  = make(keyValues, 0)
				)

				// method
				tags = append(tags, []KeyValue{
					{"method", r.Method},
				}...)

				// status code
				if rw, ok := w.(WrapResponseWriter); ok {
					tags = append(
						tags,
						KeyValue{"status_code", strconv.Itoa(rw.Status())},
					)
				}

				h, ok := histograms[label]
				if !ok {
					h = provider.NewHistogram(label, 1)
				}

				h.With(tags.tags()...).Observe(float64(time.Since(start).Milliseconds()))
			}()

			next.ServeHTTP(w, r)
		})
	}
}
