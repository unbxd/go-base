package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/unbxd/go-base/log"
)

// TraceLoggingFilter supersedes `NewTraceLoggerFinalizerHandlerOption` as this
// is more closer to the end of request handling phase.
// This reads most of the properties from Context and writes log line for loggers
// to consume.
func TraceLoggingFilter(logger log.Logger) Filter {
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

				fields = append(fields, log.Int("status", ww.Status()))

				ctx := r.Context()
				for k, ck := range map[string]ContextKey{
					"method":        ContextKeyRequestMethod,
					"proto":         ContextKeyRequestProto,
					"host":          ContextKeyRequestHost,
					"remoteAddr":    ContextKeyRequestRemoteAddr,
					"xForwardedFor": ContextKeyRequestXForwardedFor,
					"requestId":     ContextKeyRequestXRequestID,
				} {
					val := ctx.Value(ck)
					if val != nil {
						str := val.(string)

						fields = append(fields, log.String(k, str))
					}
				}

				fmt.Println(time.Since(start).Milliseconds())

				fields = append(fields, log.String("latencys", time.Since(start).String()))
				fields = append(fields, log.Int64("latency", time.Since(start).Milliseconds()))

				// trace is same as info here
				logger.Info(r.URL.RequestURI(), fields...)
			}(start)
			next.ServeHTTP(w, r)
		})
	}
}
