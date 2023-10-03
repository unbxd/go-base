package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Deprecated: use URLParser instead of URLParamsParser
type (
	URLParamsParser interface {
		ByName(string) string
	}
	chiURLParamsParser struct {
		*chi.RouteParams
	}
)

// ByName returns the URL Parameter by Name, also should implement URLParamsParser
// Deprecated: use Mux().URLParser()
func (p *chiURLParamsParser) ByName(key string) string {
	for k := len(p.Keys) - 1; k >= 0; k-- {
		if p.Keys[k] == key {
			return p.Values[k]
		}
	}
	return ""
}

// Parameters returns the request parameters extracted from
// http.Request. This method will only work with DefaultMux used in go-base
// This method here is only for backwards compatibility. Don't use this method
// Deprecated: in favour of parser := transport.Mux().URLParser(); parser.Parse(r).Get("key")
func Parameters(r *http.Request) URLParamsParser {
	return &chiURLParamsParser{
		&chi.RouteContext(r.Context()).URLParams,
	}
}
