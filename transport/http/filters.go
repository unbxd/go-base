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

// serverNameFilter is simple filter to set custom 'server' header for response
func serverNameFilter(name string, version string) Filter {
	sn := name + "-" + version
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("server", sn)
			next.ServeHTTP(w, r)
		})
	}
}

// closerFilter is builtin that wraps filter chain
func closerFilter() Filter {
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

// decorateContextFilter decorates the http.Request.Context() with
// details about the http Request
// List of keys can be found in http.go
func decorateContextFilter() Filter {
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

// NoopFilter doesn't do anything
func NoopFilter() Filter {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}
}

func heartbeatFilter(name string, heartbeats []string) Filter {
	paths := make(map[string]struct{}, len(heartbeats))
	for _, hb := range heartbeats {
		paths[hb] = struct{}{}
	}

	message := name + " :: Ah, ha, ha, ha, stayin' alive, stayin' alive!"

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodGet {
					_, ok := paths[r.URL.Path]
					if ok {
						w.Header().Set("Content-Type", "text/plain")
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(message))
						return
					}
				}
				next.ServeHTTP(w, r)
			})
	}
}
