package http

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type Filter func(http.Handler) http.Handler

func Chain(inner http.Handler, filters ...Filter) http.Handler {
	l := len(filters)
	if l == 0 {
		return inner
	}
	return filters[0](Chain(inner, filters[1:]...))
}

func CloserFilter() Filter {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				r.Body.Close()
				r.Close = true
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func RequestIDFilter(headers ...string) Filter {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			switch {
			case r.Header.Get(HeaderRequestID) != "" && len(headers) == 0:
				next.ServeHTTP(w, r)
				w.Header().Set(HeaderRequestID, r.Header.Get(HeaderRequestID))
				return
			case r.Header.Get(HeaderRequestID) != "" && len(headers) > 0:
				id := r.Header.Get(HeaderRequestID)
				for _, h := range headers {
					r.Header.Set(h, id)
				}

				next.ServeHTTP(w, r)

				w.Header().Set(HeaderRequestID, id)
				for _, h := range headers {
					w.Header().Set(h, id)
				}
				return
			default:
				id := uuid.NewString()

				r.Header.Set(HeaderRequestID, id)
				for _, h := range headers {
					r.Header.Set(h, id)
				}

				next.ServeHTTP(w, r)

				w.Header().Set(HeaderRequestID, id)
				for _, h := range headers {
					w.Header().Set(h, id)
				}
				return
			}
		})
	}
}

func DecorateContextFilter() Filter {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			ctx := r.Context()
			ctx = decorateContext(ctx, r)

			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

// CustomRequestIDFilter returns a HandlerOption for a customheader to be populated
// with request id, generated at filter
// Each Request by default should have `x-request-id` as it has been made
// default of the Transport as a filter, this is only to set the same
// value to different headers with a prefix & suffix
func CustomRequestIDFilter(prefix, suffix string, customHeaders ...string) Filter {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			id := r.Header.Get(HeaderRequestID)
			if id == "" {
				panic("failed to get request id, this shouldn't happen")
			}

			var idsb strings.Builder

			if prefix != "" {
				idsb.WriteString(prefix)
				idsb.WriteRune('-')
			}
			idsb.WriteString(id)
			if suffix != "" {
				idsb.WriteRune('-')
				idsb.WriteString(suffix)
			}

			for _, ch := range customHeaders {
				r.Header.Set(ch, idsb.String())
			}

			next.ServeHTTP(w, r)

			for _, ch := range customHeaders {
				w.Header().Set(ch, idsb.String())
			}
		})
	}
}
