package main

import (
	"context"
	net_http "net/http"

	"github.com/unbxd/go-base/base/transport/http"
)

func main() {
	tr1, _ := http.NewTransport(
		"0.0.0.0",
		"4444",
		http.WithFullDefaults(),
	)

	tr2, _ := http.NewTransport(
		"0.0.0.0",
		"4445",
		http.WithFullDefaults(),
	)

	tr1.Get("/ping", func(
		ctx context.Context, req *net_http.Request,
	) (res *net_http.Response, err error) {
		res = http.NewResponse(
			req,
			http.ResponseWithCode(net_http.StatusOK),
			http.ResponseWithBytes(
				[]byte("tr1: "+req.URL.RequestURI()),
			),
		)
		return
	})

	tr2.Get("/ping", func(
		ctx context.Context, req *net_http.Request,
	) (res *net_http.Response, err error) {
		res = http.NewResponse(
			req,
			http.ResponseWithCode(net_http.StatusOK),
			http.ResponseWithBytes(
				[]byte("tr2: "+req.URL.RequestURI()),
			),
		)
		return
	})

	go tr1.Open()
	go tr2.Open()

	for {
	}
}
