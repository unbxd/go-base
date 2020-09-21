package http

import (
	"context"
	kit_http "github.com/go-kit/kit/transport/http"
	"net/http"
)

type BeforeFunc kit_http.RequestFunc

// NoopBefore is Before which has no operation
func NoopBefore(cx context.Context, _ *http.Request) context.Context { return cx }

// NewBeforeDecorator is Before that is called before every request
// is processed by the endpoint
// The decorations done here are used by the finalizer to read the
// context and take options on it
func NewBeforeDecorator(cx context.Context, req *http.Request) context.Context {
	return kit_http.PopulateRequestContext(cx, req)
}
