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
		if cx == nil {
			return false
		}
		return true
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
		sb.WriteRune('-')
	}

	sb.WriteString(r.Method)
	sb.WriteRune(' ')
	sb.WriteString(rpt)

	return sb.String()
}

func OpenTelemetryFilterForDefaultMux(
	monitors []string,
	labels map[string]string,
) Filter {
	formatter := defaultSpanNameFormatter
	attribs := make([]attribute.KeyValue, 0)

	for k, v := range labels {
		attribs = append(attribs, attribute.String(k, v))
	}

	return otelhttp.NewMiddleware(
		"",
		otelhttp.WithFilter(filterMonitors(monitors)),
		otelhttp.WithFilter(filterDefaultMux()),
		otelhttp.WithSpanNameFormatter(formatter),
		otelhttp.WithSpanOptions(
			trace.WithNewRoot(),
			trace.WithAttributes(attribs...),
		),
	)
}
