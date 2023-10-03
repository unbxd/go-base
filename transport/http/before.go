package http

import (
	"context"
	"net/http"

	kit_http "github.com/go-kit/kit/transport/http"
)

type BeforeFunc kit_http.RequestFunc

// NoopBefore is Before which has no operation
func NoopBefore(cx context.Context, _ *http.Request) context.Context { return cx }
