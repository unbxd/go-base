package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	gorilla_mux "github.com/gorilla/mux"
)

// Muxer defines the standard Multiplexer for http Request
type (
	Muxer interface {
		// ServeHTTP
		http.Handler

		// URLParser returns the parsing method used by
		// the multiplexer
		URLParser() URLParser

		// Default Handler Method is all that is needed
		Handler(method, url string, fn http.Handler)
	}
)

type (
	URLParams map[string]string
	URLParser interface {
		Parse(r *http.Request) URLParams
	}
)

func (up URLParams) Param(key string) (value string, found bool) {
	value, found = up[key]
	return
}

func (up URLParams) ByName(key string) (value string) {
	return up[key]
}

// go-chi muxer which is default multiplexer for go-base
type (
	chiMuxer     struct{ *chi.Mux }
	chiURLParser struct{}
	ChiMuxOption func(*chiMuxer)
)

func (cup *chiURLParser) Parse(r *http.Request) URLParams {
	rcx := chi.RouteContext(r.Context())
	par := make(URLParams)

	for ix, key := range rcx.URLParams.Keys {
		par[key] = rcx.URLParams.Values[ix]
	}

	return par
}

func (mx *chiMuxer) URLParser() URLParser {
	return &chiURLParser{}
}

func (mx *chiMuxer) Handler(method, url string, fn http.Handler) {
	mx.Method(method, url, fn)
}

func newChiMux(opts ...ChiMuxOption) Muxer {
	mx := &chiMuxer{chi.NewMux()}
	for _, o := range opts {
		o(mx)
	}
	return mx
}

// gorilla mux implementation
// this here is given as an example on how to use a custom mux
// the expectation is that the application will provide its own multiplexer
// if required and not the default chi multiplexer

type (
	gmux          struct{ *gorilla_mux.Router }
	gmuxURLParser struct{}
)

func (gu *gmuxURLParser) Parse(r *http.Request) URLParams {
	return gorilla_mux.Vars(r)
}

func (gm *gmux) URLParser() URLParser {
	return &gmuxURLParser{}
}

func (gm *gmux) Handler(method, url string, fn http.Handler) {
	gm.Router.Handle(url, fn).Methods(method)
}

func NewGorillaMux() Muxer { return &gmux{gorilla_mux.NewRouter()} }
