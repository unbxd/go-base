package http

import (
	"net/http"
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
