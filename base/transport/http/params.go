package http

import (
	"net/http"

	tmux "github.com/dimfeld/httptreemux/v5"
)

// Params is wrapper on top of params in Context
type Params map[string]string

// ByName returns the URL Parameter by Name
func (p Params) ByName(name string) string { return p[name] }

// Parameters returns the request parameters extracted from
// http.Request
func Parameters(r *http.Request) Params {
	return tmux.ContextParams(r.Context())
}
