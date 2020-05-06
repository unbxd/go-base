package http

import (
	net_http "net/http"

	base_http "github.com/vtomar01/go-base/base/transport/http"
)

type gomux struct {
	*net_http.ServeMux
}

type gomuxHandler struct {
	method string
	hn     net_http.Handler
}

func (gh *gomuxHandler) ServeHTTP(rw net_http.ResponseWriter, req *net_http.Request) {
	if req.Method != gh.method {
		net_http.Error(rw, "method not allowed", net_http.StatusMethodNotAllowed)
		return
	}

	gh.hn.ServeHTTP(rw, req)
}

func (g *gomux) Handler(method, url string, hn net_http.Handler) {
	g.Handle(url, &gomuxHandler{method, hn})
}

// NewGomux returns the default Golang Multiplexer implementation
func NewGomux() base_http.Mux {
	return &gomux{
		net_http.NewServeMux(),
	}
}
