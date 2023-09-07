package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Params is wrapper on top of params in Context
type Params chi.RouteParams

// ByName returns the URL Parameter by Name
func (p Params) ByName(key string) string {
	for k := len(p.Keys) - 1; k >= 0; k-- {
		if p.Keys[k] == key {
			return p.Values[k]
		}
	}
	return ""
}

// Parameters returns the request parameters extracted from
// http.Request
func Parameters(r *http.Request) Params {
	return Params(chi.RouteContext(r.Context()).URLParams)
}
