package main

import (
	"context"
	"errors"
	clog "log"
	"math/rand"
	net_http "net/http"

	"github.com/unbxd/go-base/base/log"
	"github.com/unbxd/go-base/base/transport/http"
)

var (
	errOne   = errors.New("New Error 1")
	errTwo   = errors.New("New Error 2")
	errThree = errors.New("New Error 3")
	errFour  = errors.New("New Error 4")
	errFive  = errors.New("New Error 5")

	errs = []error{
		errOne, errTwo, errThree, errFour, errFive,
	}
)

func errEncoder(ctx context.Context, err error, w net_http.ResponseWriter) {
	switch err {
	case errOne:
		w.WriteHeader(net_http.StatusNotAcceptable)
		w.Write([]byte(
			"ERROR is Error 1 : " + errOne.Error() +
				" Status: " + net_http.StatusText(net_http.StatusNotAcceptable),
		))
	case errTwo:
		w.WriteHeader(net_http.StatusMethodNotAllowed)
		w.Write([]byte(
			"ERROR is Error 2 : " + errTwo.Error() +
				" Status: " + net_http.StatusText(net_http.StatusMethodNotAllowed),
		))
	case errThree:
		w.WriteHeader(net_http.StatusInternalServerError)
		w.Write([]byte(
			"ERROR is Error 3 : " + errThree.Error() +
				" Status: " + net_http.StatusText(net_http.StatusInternalServerError),
		))
	case errFour:
		w.WriteHeader(net_http.StatusNotFound)
		w.Write([]byte(
			"ERROR is Error 4 : " + errFour.Error() +
				" Status: " + net_http.StatusText(net_http.StatusNotFound),
		))
	default:
		w.WriteHeader(net_http.StatusOK)
		w.Write([]byte("all good"))

	}
}

func main() {
	logger, err := log.NewZapLogger(
		"debug", "console", []string{"stdout"},
	)
	if err != nil {
		clog.Fatal("Error init logging", err)
	}

	tr, err := http.NewTransport(
		"0.0.0.0",
		"4444",
		http.WithMonitors([]string{"/health_check.html"}),
		http.WithLogger(logger),
		http.WithFullDefaults(),
		http.WithErrorEncoder(errEncoder),
	)

	if err != nil {
		clog.Fatal("Error init server:", err)
	}
	/*
		This is a simple example of a request being handled.
	*/
	tr.Get("/hello-world", func(
		ctx context.Context,
		req *net_http.Request,
	) (res *net_http.Response, err error) {
		return http.NewByteResponse(req, []byte("hello-world")), err
	})

	/*
		By default to handle Error, kit/transport/http.DefaultErrorEncoder is used

		To handle Custom Error Conditions

		- Write Custom Errors
		- Write Custom ErrorEncoder
		- use http.WithErrorEncoder() to bind Error Encoders

		To Have custom Error Encoder, use the ServerOption that can be
		passed using `transport.Get(url, fn, [...ServerOption])` <- these Server Options
	*/

	tr.Get("/error", func(
		ctx context.Context,
		req *net_http.Request,
	) (res *net_http.Response, err error) {
		num := rand.Intn(4)
		return nil, errs[num]
	})

	tr.Open()
}
