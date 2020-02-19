package http

import (
	"context"
	"net/http"
	net_http "net/http"

	"github.com/julienschmidt/httprouter"
)

// decorateEndpoint returns endpoint.Endpoint by wrapping around HandlerFunc
func decorateEndpoint(fn HandlerFunc) Handler {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		rr, ok := req.(*net_http.Request)
		if !ok {
			return nil, ErrNotHTTPRequest
		}
		return fn(ctx, rr)
	}
}

func encapsulate(
	fn HandlerFunc,
	trs []HandlerOption,
	pats []HandlerOption,
) net_http.Handler {
	return newHandler(
		decorateEndpoint(fn),
		append(trs, pats...)...,
	)
}

// Params is wrapper on top of httprouter.Param
type Params struct {
	httprouter.Params
}

// ByName returns the URL Parameter by Name
func (p *Params) ByName(name string) string { return p.Params.ByName(name) }

// Parameters returns the request parameters extracted from
// http.Request
func Parameters(r *http.Request) Params {
	return Params{httprouter.ParamsFromContext(r.Context())}
}
