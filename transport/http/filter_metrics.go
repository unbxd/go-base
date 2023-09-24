package http

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

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

// filters if we don't use default multiplexer
func filterDefaultMux() otelhttp.Filter {
	return func(r *http.Request) bool {
		cx := chi.RouteContext(r.Context())
		return cx != nil
	}
}

type SpanNameFormatter func(operation string, r *http.Request) string

func defaultSpanNameFormatter(operation string, r *http.Request) string {
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

// OpenTelemetryFilterForDefaultMux uses OpenTelemetry to publish events
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
func OpenTelemetryFilterForDefaultMux(
	monitors []string,
	labels map[string]string,
) Filter {
	formatter := defaultSpanNameFormatter
	attribs := make([]attribute.KeyValue, 0)

	for k, v := range labels {
		attribs = append(attribs, attribute.String(k, v))
	}

	// this is slightly in-efficient that we are double wrapping
	// http.ResponseWriter.
	// In this middleware, there is a wrapping of 'respWriterWrapper'
	// to extract the status code, which we do too in case of
	// `WrapResponseWriter`
	// This in itself shouldn't cause any issue because ResponseWriter
	// is an interface and we are wrapping it twice, just that
	// it introduces a bit of overhead of computation
	return otelhttp.NewMiddleware(
		"http-serve",
		otelhttp.WithFilter(filterMonitors(monitors)),
		otelhttp.WithFilter(filterDefaultMux()),
		otelhttp.WithSpanNameFormatter(formatter),
		otelhttp.WithSpanOptions(
			trace.WithNewRoot(),
			trace.WithAttributes(attribs...),
		),
	)
}
