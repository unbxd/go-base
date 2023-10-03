package http

import (
	"net/http"
	net_http "net/http"
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

// SetHeaderFilter is a convenience handler to set a response header key/value
func SetHeaderFilter(key, value string) Filter { return middleware.SetHeader(key, value) }

// SetRequestHeader sets custom header for downstream to consume, useful if we
// call other downstreams directly from http layer
func SetRequestHeader(key, value string) Filter {
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
		AllowedMethods:   []string{net_http.MethodGet, net_http.MethodHead, net_http.MethodOptions},
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
	return func(next net_http.Handler) net_http.Handler {
		return http.HandlerFunc(
			func(w net_http.ResponseWriter, r *net_http.Request) {
				for _, h := range headers {
					r.Header.Del(h)
				}

				next.ServeHTTP(w, r)
			},
		)
	}
}
