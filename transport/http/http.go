package http

import (
	"context"
	net_http "net/http"

	"github.com/unbxd/go-base/v2/errors"
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

func decorateContext(ctx context.Context, r *net_http.Request) context.Context {
	for k, v := range map[ContextKey]string{
		ContextKeyRequestMethod:          r.Method,
		ContextKeyRequestURI:             r.RequestURI,
		ContextKeyRequestPath:            r.URL.Path,
		ContextKeyRequestProto:           r.Proto,
		ContextKeyRequestHost:            r.Host,
		ContextKeyRequestRemoteAddr:      r.RemoteAddr,
		ContextKeyRequestXForwardedFor:   r.Header.Get(HeaderXForwardedFor),
		ContextKeyRequestXForwardedProto: r.Header.Get(HeaderXForwardedProto),
		ContextKeyRequestAuthorization:   r.Header.Get(HeaderAuthorization),
		ContextKeyRequestReferer:         r.Header.Get(HeaderReferer),
		ContextKeyRequestUserAgent:       r.Header.Get(HeaderUserAgent),
		ContextKeyRequestXRequestID:      r.Header.Get(HeaderRequestID),
		ContextKeyRequestAccept:          r.Header.Get(HeaderAccept),
	} {
		ctx = context.WithValue(ctx, k, v)
	}
	return ctx
}

// Headers
const (
	HeaderAllowHeaders    = "Access-Control-Allow-Headers"
	HeaderAllowMethods    = "Access-Control-Allow-Methods"
	HeaderAllowOrigin     = "Access-Control-Allow-Origin"
	HeaderExposeHeader    = "Access-Control-Expose-Headers"
	HeaderAccessMaxAge    = "Access-Control-Max-Age"
	HeaderRequestID       = "X-Request-Id"
	HeaderXForwardedFor   = "X-Forwarded-For"
	HeaderXForwardedProto = "X-Forwarded-Proto"
	HeaderAuthorization   = "Authorization"
	HeaderReferer         = "Referer"
	HeaderUserAgent       = "User-Agent"
	HeaderAccept          = "Accept"
	HeaderContentType     = "Content-Type"
)

// Standard HTTP Errors
var (
	ErrNotHTTPRequest  = errors.New("missmatch type, request should be *http.Request")
	ErrNotHTTPResponse = errors.New("mismatch type, response should be *http.Response")
)
