package http

import (
	"net/http"

	"github.com/google/uuid"
)

func requestIDFilter(headers ...string) Filter {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			switch {
			// case where requestId is sent by upstream
			case r.Header.Get(HeaderRequestID) != "" && len(headers) == 0:
				w.Header().Set(HeaderRequestID, r.Header.Get(HeaderRequestID))
				next.ServeHTTP(w, r)
			// case where requestId is sent by upstream and we write bunch of other headers
			case r.Header.Get(HeaderRequestID) != "" && len(headers) > 0:
				id := r.Header.Get(HeaderRequestID)
				for _, h := range headers {
					r.Header.Set(h, id)
				}
				w.Header().Set(HeaderRequestID, id)
				for _, h := range headers {
					w.Header().Set(h, id)
				}
				next.ServeHTTP(w, r)
			// case where we need to generate requestId
			default:
				id := uuid.NewString()

				r.Header.Set(HeaderRequestID, id)
				for _, h := range headers {
					r.Header.Set(h, id)
				}

				w.Header().Set(HeaderRequestID, id)

				for _, h := range headers {
					w.Header().Set(h, id)
				}

				next.ServeHTTP(w, r)
			}
		})
	}
}

type RequestIDFormatter func(uuid string) string

// CustomRequestIDFilter returns a HandlerOption for a customheader to be populated
// with request id, generated at filter
// Each Request by default should have `x-request-id` as it has been made
// default of the Transport as a filter, this is only to set the same
// value to different headers with a prefix & suffix
func CustomRequestIDFilter(formatter RequestIDFormatter, customHeaders ...string) Filter {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			id := r.Header.Get(HeaderRequestID)
			if id == "" {
				panic("failed to get request id, this shouldn't happen")
			}

			id = formatter(id)

			for _, ch := range customHeaders {
				r.Header.Set(ch, id)
			}

			for _, ch := range customHeaders {
				w.Header().Set(ch, id)
			}

			next.ServeHTTP(w, r)
		})
	}
}
