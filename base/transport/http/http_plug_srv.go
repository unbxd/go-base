package http

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-kit/kit/endpoint"

	// kit_log "github.com/go-kit/kit/log"
	"net/http"
	net_http "net/http"

	"github.com/go-kit/kit/metrics"
	kit_http "github.com/go-kit/kit/transport/http"
	uuid "github.com/satori/go.uuid"
	"github.com/uknth/go-base/base/log"
)

// ContextKey is key for context
type ContextKey int

// ContextKeys
const (
	ContextKeyRequestMethod ContextKey = iota
	ContextKeyRequestURI
	ContextKeyRequestPath
	ContextKeyRequestProto
	ContextKeyRequestHost
	ContextKeyRequestRemoteAddr
	ContextKeyRequestXForwardedFor
	ContextKeyRequestXForwardedProto
	ContextKeyRequestAuthorization
	ContextKeyRequestReferer
	ContextKeyRequestUserAgent
	ContextKeyRequestXRequestID
	ContextKeyRequestAccept
	ContextKeyResponseHeaders
	ContextKeyResponseSize
)

// NewTraceLoggerFinalizer returns a kit_http.ServerOption for simple trace logging
func NewTraceLoggerFinalizer(logger log.Logger) kit_http.ServerOption {
	return kit_http.ServerFinalizer(func(ctx context.Context, code int, r *net_http.Request) {
		// safety check if someone includes logging
		// but doesn't provide a logger
		if logger == nil {
			return
		}

		var fields = []log.Field{log.Int("code", code)}
		for k, ck := range map[string]ContextKey{
			"method":          ContextKeyRequestMethod,
			"proto":           ContextKeyRequestProto,
			"host":            ContextKeyRequestHost,
			"remote_addr":     ContextKeyRequestRemoteAddr,
			"x-forwarded-for": ContextKeyRequestXForwardedFor,
			"x-request-id":    ContextKeyRequestXRequestID,
		} {
			val := ctx.Value(ck)
			if val != nil {
				str := val.(string)

				fields = append(fields, log.String(k, str))
			}
		}

		logger.Info(r.URL.RequestURI(), fields...)
	})
}

// NewRequestIDRequestFunc returns a kit_http.ServerOption  for simple Request ID generation
func NewRequestIDRequestFunc(customHeaders ...string) kit_http.ServerOption {
	rid := "X-Request-Id"

	fn := func(ctx context.Context, r *net_http.Request) context.Context {
		// ignore if already set & no custom headers
		if r.Header.Get(rid) != "" && len(customHeaders) == 0 {
			return ctx
		}

		id := uuid.NewV4().String()

		if r.Header.Get(rid) != "" {
			id = r.Header.Get(rid)
		}

		r.Header.Set(rid, id)

		for _, hr := range customHeaders {
			r.Header.Set(hr, id)
		}

		return ctx
	}
	return kit_http.ServerBefore(fn)
}

// MetricsFunc is a callback to extract necessary tags for the metrics
// based on Request & Response recieved
type MetricsFunc func(req interface{}, res interface{}) []string

// NewHistogramInstrumentation ...
func NewHistogramInstrumentation(
	tags []string,
	histogram metrics.Histogram,
	fn MetricsFunc,
) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, req interface{}) (res interface{}, err error) {
			defer func(start time.Time) {
				if fn != nil {
					tags = append(
						tags, fn(req, res)...,
					)
				}
				tags = append(
					tags, []string{"error", fmt.Sprint(err != nil)}...,
				)
				histogram.With(tags...).Observe(time.Since(start).Seconds())
			}(time.Now())

			res, err = next(ctx, req)
			return
		}
	}
}

// NewCounterInstrumentation allows for counter to be incremented and instrumented
func NewCounterInstrumentation(
	tags []string,
	incr float64,
	counter metrics.Counter,
	fn MetricsFunc,
) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, req interface{}) (res interface{}, err error) {
			defer func(start time.Time) {
				if fn != nil {
					tags = append(
						tags, fn(req, res)...,
					)
				}
				tags = append(
					tags, []string{"error", fmt.Sprint(err != nil)}...,
				)
				counter.With(tags...).Add(incr)
			}(time.Now())

			res, err = next(ctx, req)
			return
		}
	}
}

// NewServerResponseFunc sets Server Header for the Request
func NewServerResponseFunc(appName string) kit_http.ServerOption {
	var header = appName + "-server"

	return kit_http.ServerAfter(
		func(ctx context.Context, rw net_http.ResponseWriter) context.Context {
			rw.Header().Set("Server", header)
			return ctx
		})
}

const (
	headerAllowHeaders = "Access-Control-Allow-Headers"
	headerAllowMethods = "Access-Control-Allow-Methods"
	headerAllowOrigin  = "Access-Control-Allow-Origin"
	headerExposeHeader = "Access-Control-Expose-Headers"
	headerAccessMaxAge = "Access-Control-Max-Age"
)

var (
	deforigin    = "*"
	defmaxage    = 5
	defmethods   = []string{net_http.MethodGet, net_http.MethodHead, net_http.MethodOptions}
	defheaders   = []string{"Content-Type", "Accept-Encoding", "X-Request-Id"}
	defexheaders = []string{"X-Request-Id", "trace-id"}
)

// NewCustomCORSResponseFunc sets CORS header for a given request
func NewCustomCORSResponseFunc(
	origin string,
	maxage int,
	methods []string,
	headers []string,
	exposeHeaders []string,
) kit_http.ServerOption {

	return kit_http.ServerAfter(func(
		ctx context.Context, rw net_http.ResponseWriter,
	) context.Context {

		if origin == "" {
			origin = deforigin
		}

		if maxage < 0 {
			maxage = defmaxage
		}

		if len(methods) == 0 {
			methods = defmethods
		}

		if len(headers) == 0 {
			headers = defheaders
		}

		if len(exposeHeaders) == 0 {
			exposeHeaders = defexheaders
		}

		rw.Header().Set(headerAllowOrigin, origin)
		rw.Header().Set(headerAccessMaxAge, strconv.Itoa(maxage))
		rw.Header().Set(headerAllowMethods, strings.Join(methods, ","))
		rw.Header().Set(headerAllowHeaders, strings.Join(headers, ","))
		rw.Header().Set(headerExposeHeader, strings.Join(exposeHeaders, ","))

		return ctx
	})
}

// NewCORSResponseFunc sets default CORS headers for a request
func NewCORSResponseFunc() kit_http.ServerOption {
	return kit_http.ServerAfter(
		func(ctx context.Context, rw net_http.ResponseWriter) context.Context {

			rw.Header().Set(headerAllowOrigin, deforigin)
			rw.Header().Set(headerAccessMaxAge, strconv.Itoa(defmaxage))
			rw.Header().Set(headerAllowMethods, strings.Join(defmethods, ","))
			rw.Header().Set(headerAllowHeaders, strings.Join(defheaders, ","))
			rw.Header().Set(headerExposeHeader, strings.Join(defexheaders, ","))

			return ctx
		},
	)
}

// NewDefaultErrorEncoder sets default error encoder for endpoint
func NewDefaultErrorEncoder() kit_http.ServerOption {
	return kit_http.ServerErrorEncoder(kit_http.DefaultErrorEncoder)
}

// NewDeleteHeaderRequestFunc deletes the headers from net_http.Request
// before it is sent to HandlerFunc
func NewDeleteHeaderRequestFunc(headers ...string) kit_http.ServerOption {
	return kit_http.ServerBefore(
		func(ctx context.Context, req *net_http.Request) context.Context {

			for _, h := range headers {
				req.Header.Del(h)
			}

			return ctx
		},
	)
}

// NewPopulateRequestContextRequestFunc populates the context with
// properties extracted from net_http.Request
func NewPopulateRequestContextRequestFunc() kit_http.ServerOption {
	return kit_http.ServerBefore(
		func(ctx context.Context, r *http.Request) context.Context {
			for k, v := range map[ContextKey]string{
				ContextKeyRequestMethod:          r.Method,
				ContextKeyRequestURI:             r.RequestURI,
				ContextKeyRequestPath:            r.URL.Path,
				ContextKeyRequestProto:           r.Proto,
				ContextKeyRequestHost:            r.Host,
				ContextKeyRequestRemoteAddr:      r.RemoteAddr,
				ContextKeyRequestXForwardedFor:   r.Header.Get("X-Forwarded-For"),
				ContextKeyRequestXForwardedProto: r.Header.Get("X-Forwarded-Proto"),
				ContextKeyRequestAuthorization:   r.Header.Get("Authorization"),
				ContextKeyRequestReferer:         r.Header.Get("Referer"),
				ContextKeyRequestUserAgent:       r.Header.Get("User-Agent"),
				ContextKeyRequestXRequestID:      r.Header.Get("X-Request-Id"),
				ContextKeyRequestAccept:          r.Header.Get("Accept"),
			} {
				ctx = context.WithValue(ctx, k, v)
			}
			return ctx
		},
	)
}

// NewSetRequestHeader sets the request header
func NewSetRequestHeader(key, val string) kit_http.ServerOption {
	return kit_http.ServerBefore(kit_http.SetRequestHeader(key, val))
}

// NewSetResponseHeader sets the response header
func NewSetResponseHeader(key, val string) kit_http.ServerOption {
	return kit_http.ServerAfter(kit_http.SetResponseHeader(key, val))
}
