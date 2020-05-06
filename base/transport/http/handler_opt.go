package http

import (
	"context"
	"strconv"
	"strings"

	"net/http"
	net_http "net/http"

	kit_http "github.com/go-kit/kit/transport/http"
	uuid "github.com/satori/go.uuid"
	"github.com/vtomar01/go-base/base/log"
)

// NewTraceLoggerFinalizerHandlerOption returns a HandlerOption for simple trace logging
func NewTraceLoggerFinalizerHandlerOption(logger log.Logger) HandlerOption {
	return func(h *handler) {
		option := kit_http.ServerFinalizer(
			func(ctx context.Context, code int, r *net_http.Request) {
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
			},
		)
		h.options = append(h.options, option)
	}
}

// NewRequestIDHandlerOption returns a HandlerOption for simple Request ID generation
func NewRequestIDHandlerOption(customHeaders ...string) HandlerOption {
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
		ctx = context.WithValue(ctx, ContextKeyRequestXRequestID, id)

		for _, hr := range customHeaders {
			r.Header.Set(hr, id)
		}

		return ctx
	}

	return func(h *handler) {
		h.options = append(h.options, kit_http.ServerBefore(fn))
	}
}

// NewServerHandlerOption sets Server Header for the Request
func NewServerHandlerOption(name, version string) HandlerOption {
	var header = name + "-server:" + version

	return func(h *handler) {
		h.options = append(h.options, kit_http.ServerAfter(
			func(ctx context.Context, rw net_http.ResponseWriter) context.Context {
				rw.Header().Set("Server", header)
				return ctx
			}),
		)
	}
}

var (
	deforigin    = "*"
	defmaxage    = 5
	defmethods   = []string{net_http.MethodGet, net_http.MethodHead, net_http.MethodOptions}
	defheaders   = []string{"Content-Type", "Accept-Encoding", "X-Request-Id"}
	defexheaders = []string{"X-Request-Id", "trace-id"}
)

// NewCustomCORSHandlerOption sets CORS header for a given request
func NewCustomCORSHandlerOption(
	origin string,
	maxage int,
	methods []string,
	headers []string,
	exposeHeaders []string,
) HandlerOption {

	return func(h *handler) {
		h.options = append(h.options, kit_http.ServerAfter(
			func(
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

				rw.Header().Set(HeaderAllowOrigin, origin)
				rw.Header().Set(HeaderAccessMaxAge, strconv.Itoa(maxage))
				rw.Header().Set(HeaderAllowMethods, strings.Join(methods, ","))
				rw.Header().Set(HeaderAllowHeaders, strings.Join(headers, ","))
				rw.Header().Set(HeaderExposeHeader, strings.Join(exposeHeaders, ","))

				return ctx
			}),
		)
	}
}

// NewCORSHandlerOption sets default CORS headers for a request
func NewCORSHandlerOption() HandlerOption {
	return func(h *handler) {
		h.options = append(h.options, kit_http.ServerAfter(
			func(ctx context.Context, rw net_http.ResponseWriter) context.Context {
				rw.Header().Set(HeaderAllowOrigin, deforigin)
				rw.Header().Set(HeaderAccessMaxAge, strconv.Itoa(defmaxage))
				rw.Header().Set(HeaderAllowMethods, strings.Join(defmethods, ","))
				rw.Header().Set(HeaderAllowHeaders, strings.Join(defheaders, ","))
				rw.Header().Set(HeaderExposeHeader, strings.Join(defexheaders, ","))

				return ctx
			},
		))
	}
}

// NewDeleteHeaderHandlerOption deletes the headers from net_http.Request
// before it is sent to HandlerFunc
func NewDeleteHeaderHandlerOption(headers ...string) HandlerOption {
	return func(h *handler) {
		h.options = append(h.options, kit_http.ServerBefore(
			func(ctx context.Context, req *net_http.Request) context.Context {

				for _, h := range headers {
					req.Header.Del(h)
				}

				return ctx
			},
		))
	}
}

func populateRequestContext(ctx context.Context, r *http.Request) context.Context {
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
}

// NewPopulateRequestContextRequestFunc populates the context with
// properties extracted from net_http.Request
func NewPopulateRequestContextRequestFunc() HandlerOption {
	return func(h *handler) {
		h.options = append(h.options, kit_http.ServerBefore(
			populateRequestContext,
		))
	}
}

// NewSetRequestHeader sets the request header
func NewSetRequestHeader(key, val string) HandlerOption {
	return func(h *handler) {
		h.options = append(
			h.options,
			kit_http.ServerBefore(kit_http.SetRequestHeader(key, val)),
		)
	}
}

// NewSetResponseHeader sets the response header
func NewSetResponseHeader(key, val string) HandlerOption {
	return func(h *handler) {
		h.options = append(
			h.options,
			kit_http.ServerAfter(kit_http.SetResponseHeader(key, val)),
		)
	}
}
