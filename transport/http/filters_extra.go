package http

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// collection of filters to be used with transport or handlers
// AllowContentTypeFilter enforces a whitelist of request Content-Types otherwise responds
// with a 415 Unsupported Media Type status.
func AllowContentTypeFilter(contentTypes ...string) Filter {
	allowedContentTypes := make(map[string]struct{}, len(contentTypes))
	for _, ctype := range contentTypes {
		allowedContentTypes[strings.TrimSpace(strings.ToLower(ctype))] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if r.ContentLength == 0 {
				// skip check for empty content body
				next.ServeHTTP(w, r)
				return
			}

			s := strings.ToLower(strings.TrimSpace(r.Header.Get("Content-Type")))
			if i := strings.Index(s, ";"); i > -1 {
				s = s[0:i]
			}

			if _, ok := allowedContentTypes[s]; ok {
				next.ServeHTTP(w, r)
				return
			}

			w.WriteHeader(http.StatusUnsupportedMediaType)
		}
		return http.HandlerFunc(fn)
	}
}

// GZipCompressionFilter is a middleware that compresses response body of a given content types to a data format based on Accept-Encoding request header. It uses a given compression level.
// NOTE: make sure to set the Content-Type header on your response otherwise this middleware will not compress the response body. For ex, in your handler you should set w.Header().Set("Content-Type", http.DetectContentType(yourBody)) or set it manually.
// Passing a compression level of 5 is sensible value
func GzipCompressionFilter(level int, types ...string) Filter {
	return middleware.Compress(level, types...)
}

// CleanPath middleware will clean out double slash mistakes from a user's request path. For example, if a user requests /users//1 or //users////1 will both be treated as: /users/1
func CleanPathFilter() Filter {
	return middleware.CleanPath
}

// RedirectSlashes is a middleware that will match request paths with a trailing
// slash and redirect to the same path, less the trailing slash.
//
// NOTE: RedirectSlashes middleware is *incompatible* with http.FileServer,
// see https://github.com/go-chi/chi/issues/343
func RedirectSlashFilter() Filter { return middleware.RedirectSlashes }

// SetResponseHeaderFilter is a convenience handler to set a response header key/value
func SetResponseHeaderFilter(key, value string) Filter { return middleware.SetHeader(key, value) }

// SetRequestHeaderFilter sets custom header for downstream to consume, useful if we
// call other downstreams directly from http layer
func SetRequestHeaderFilter(key, value string) Filter {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Header.Set(key, value)
			next.ServeHTTP(w, r)
		})
	}
}

// CorsFilterWithDefaults is cors filter with sane defaults. It sets the
// https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS
// headers for any server response
func CorsFilterWithDefaults() Filter {
	opts := cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{http.MethodGet, http.MethodHead, http.MethodOptions},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "Accept-Encoding", "X-Request-ID"},
		ExposedHeaders:   []string{"X-Request-Id", "trace-id", "Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}

	return cors.Handler(opts)
}

type CorsOptions cors.Options

// CorsFilterWithCustomOptions is cors filter with custom configurable options
func CorsFilterWithCustomOptions(options CorsOptions) Filter {
	return cors.Handler(cors.Options(options))
}

// DeleteHadersFilter deletes the headers in http request
func DeleteHeadersFilter(headers ...string) Filter {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				for _, h := range headers {
					r.Header.Del(h)
				}

				next.ServeHTTP(w, r)
			},
		)
	}
}

// WithChi allows you to use chi's supported middleware.
// To see the list of available middlewares, refer https://github.com/go-chi/chi?tab=readme-ov-file#core-middlewares
// For example, to use middleware.BasicAuth, set it as follows
//
//		transportConfigOptions := append(
//			transportConfigOptions,
//			[]http.TransportConfigOption{
//				http.WithCustomHostPort(cx.String("http.host"), cx.String("http.port")),
//				http.WithTraceLogging(),
//				http.WithFilters(
//					WithChi(middleware.BasicAuth(...)),
//				),
//			}...,
//		),
//	)
func WithChi(middleware func(http.Handler) http.Handler) Filter {
	return middleware
}

// ---- internal filters, shouldn't be used externally ----

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

// noopFilter is no-action filter
func noopFilter() Filter {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
			ctx := decorateContext(r.Context(), r)
			next.ServeHTTP(w, r.WithContext(ctx))
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
				if r.Method == http.MethodGet || r.Method == http.MethodHead {
					_, ok := paths[r.URL.Path]
					if ok {
						w.Header().Set("Content-Type", "text/plain")
						w.WriteHeader(http.StatusOK)
						_, _ = w.Write([]byte(message))
						return
					}
				}
				next.ServeHTTP(w, r)
			})
	}
}
