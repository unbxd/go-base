package http

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

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
