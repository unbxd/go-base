package http

import (
	"net/http"
	"time"

	"github.com/unbxd/go-base/v2/log"
)

func nc(val interface{}) string {
	if val == nil {
		return "nil"
	}

	st, ok := val.(string)
	if !ok {
		return "notstr"
	}
	return st
}

type TraceLogFieldsGen func(rw WrapResponseWriter, req *http.Request) []log.Field

// TraceLoggingFilter supersedes `NewTraceLoggerFinalizerHandlerOption` as this
// is more closer to the end of request handling phase.
// This reads most of the properties from Context and writes log line for loggers
// to consume.
func TraceLoggingFilter(logger log.Logger, fieldGenerators ...TraceLogFieldsGen) Filter {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			defer func(start time.Time) {
				// calculate fields
				fields := make([]log.Field, 0, 10)
				// status-code
				ww, ok := w.(WrapResponseWriter)
				if !ok {
					// this should never happen, panic
					msg := `
					responseWriter is not 'WrapResponseWriter'.
					did you miss 'WrappedResponseFilter' in default definition"
					`
					panic(msg)
				}

				ctx := r.Context()
				for k, ck := range map[string]ContextKey{
					"req.method":      ContextKeyRequestMethod,
					"req.uri":         ContextKeyRequestURI,
					"req.path":        ContextKeyRequestPath,
					"req.host":        ContextKeyRequestHost,
					"req.remote_addr": ContextKeyRequestRemoteAddr,
					"req.xfor":        ContextKeyRequestXForwardedFor,
					"req.ref":         ContextKeyRequestReferer,
					"req.id":          ContextKeyRequestXRequestID,
					"req.hdr.accept":  ContextKeyRequestAccept,
					"res.size":        ContextKeyResponseSize,
				} {
					fields = append(fields, log.String(k, nc(ctx.Value(ck))))
				}

				fields = append(fields, log.Int("status", ww.Status()))

				for _, fg := range fieldGenerators {
					fields = append(fields, fg(ww, r)...)
				}

				end := time.Since(start)

				fields = append(fields, log.String("latencys", end.String()))
				fields = append(fields, log.Int64("latency", end.Milliseconds()))

				// trace is same as info here
				logger.Info(r.URL.RequestURI(), fields...)
			}(start)
			next.ServeHTTP(w, r)
		})
	}
}
