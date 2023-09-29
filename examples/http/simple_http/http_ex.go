package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	clog "log"
	"math/rand"
	net_http "net/http"
	"time"

	"github.com/unbxd/go-base/log"
	"github.com/unbxd/go-base/transport/http"
)

/*
	This examples demonstrates the capabilities of go-base. It specifically
	deals with HTTP transport.

	The main function written below describes the various ways in which
	a request/response can be handled.
*/

func main() {
	// Create New Logger for the sample application.
	// Typically this would be needed for any project where logging is needed.
	// Logging is provided as direct implementation of
	// log.Logger interface.
	logger, err := log.NewZeroLogger(
		log.ZeroLoggerWithLevel("debug"),
		log.ZeroLoggerWithCaller(),
	)
	if err != nil {
		clog.Fatal("error init logging", err)
	}

	// Once we have defined log, we create a transport. Given that an
	// application can have more than one Transport, like Kafka & HTTP both
	// or NATS & HTTP both, all transports can be run as separate go-routines

	// Let's define HTTP Transport
	// Transport can be defined with a function signature
	//
	// 	http.NewTransport(host, port string, opts TransportOptions ...)
	//
	// There are lots of TransportOptions defined in go-base/kit/transport/http
	// package and can be used in creation of transport.
	// A sane default could be using
	//
	// 	http.WithFullDefaults()
	//
	// function, which has bunch of Options pre defined.

	tr, err := http.NewHTTPTransport(
		"gobi-example",
		http.WithVersion("v1.0.0"),
		http.WithCustomHostPort("0.0.0.0", "4444"),
		http.WithCustomLogger(logger),
		http.WithDefaultTransportOptions(http.WithErrorEncoder(errEncoder)),
		http.WithFilters(func(handler net_http.Handler) net_http.Handler {
			return net_http.HandlerFunc(func(rw net_http.ResponseWriter, r *net_http.Request) {
				r.Header.Add("handlers-in-order-h1", "h1")
				handler.ServeHTTP(rw, r)
			})
		}),
		http.WithFilters(func(handler net_http.Handler) net_http.Handler {
			return net_http.HandlerFunc(func(rw net_http.ResponseWriter, r *net_http.Request) {
				r.Header.Add("handlers-in-order-h2", "h2")
				handler.ServeHTTP(rw, r)
			})
		}),
	)
	//
	//
	// NOTE: An application can only have one single http.Transport
	//       This is because the httpMux cannot handle the same
	//       route in two different transport.
	//       However, if you have two completely different set of
	//       routes, it should be fine.
	//
	//
	// This above steps, creates a transport which supports
	// - some monitor endpoints, which are basically just simple `ping`
	// - with transport level logging with a custom logger
	// - with sane defaults like
	// 	- default trace logging on the above logger
	//	- request id generation
	// 	- CORS
	//	- simple error handler function
	// - with custom error encoder which handles the error case
	if err != nil {
		clog.Fatal("Error init server:", err)
	}

	// Once a transport is created, various endpoints can be attached to
	// it. These endpoints are of two kinds.
	//
	//	- http.HandlerFunc
	//	- http.Handler
	//
	// http.HandlerFunc is a simple function which takes *net_http.Request
	// and returns *net_http.Response
	// It is designed to expose only those endpoints which mostly have
	// network level tasks, and abstraction. A simple example of an API
	// only using network and returning `hello-world` is defined below
	tr.Get("/hello-world",
		func(
			ctx context.Context,
			req *net_http.Request,
		) (res *net_http.Response, err error) {
			return http.NewResponse(
				req,
				http.ResponseWithBytes([]byte("hello-world")),
			), err
		},
		http.HandlerWithFilter(func(handler net_http.Handler) net_http.Handler {
			return net_http.HandlerFunc(func(rw net_http.ResponseWriter, r *net_http.Request) {
				r.Header.Add("handlers in order", "h3")
				handler.ServeHTTP(rw, r)
			})
		}),
		http.HandlerWithFilter(func(handler net_http.Handler) net_http.Handler {
			return net_http.HandlerFunc(func(rw net_http.ResponseWriter, r *net_http.Request) {
				r.Header.Add("handlers in order", "h4")
				handler.ServeHTTP(rw, r)
			})
		}),
	)

	// For cases where we need to have objects of business domain, use http.Handler
	// A typical request handled in an application has three main phases
	// 	1. Get the Request and Translate it into Business Object (B1)
	//	2. Analyse the business object (B1) and respond with a new Business Object (B2)
	//	3. Convert Business Object (B2) back to network object like http.Response
	//
	// To support such a functionality, transport exposes a set of method
	// supported by all caps functions. eg: tr.GET, or tr.POST
	// Let's take an example of tr.POST, this is assuming that we have a
	// business object sent over the wire
	//
	// Business Object (B1) aka MODEL
	type Employee struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	// to read this business object from the network object, i.e. http.Request
	// Decoder is used
	// Decoder is defined as `func(context.Context, *net_http.Request) (interface{}, error)`
	// where in the function reads the net_http.Request and translates it into B1
	decoderFunc := func(
		_ context.Context, req *net_http.Request,
	) (interface{}, error) {
		// here we read req.Body into the object Employee
		var emp Employee

		err := json.NewDecoder(req.Body).Decode(&emp)
		if err != nil {
			return nil, fmt.Errorf("Error in Decoding: %s", err.Error())
		}

		// once decodes succeeds, the Business Object B1 is read from
		// request.Body and decoded in emp
		return emp, nil
	}

	// Now Let's say we read employee object and pass it onto a function
	// which takes in employee and returns a new object Manager out
	// Let's define Manager object then
	type Manager struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	// this is the function which actually converts the Employee to Manager
	serviceFunc := func(
		_ context.Context, emp Employee,
	) (Manager, error) {
		return Manager{emp.ID + "-manager", emp.Name + "-manager"}, nil
	}

	// The final step in this whole flow would be then to convert Manager to
	// corresponding JSON. This step is called Encoding, and is done by http.Encoder
	encodingFunc := func(_ context.Context, rw net_http.ResponseWriter, b2 interface{}) error {
		// b2 is the business object mentioned above.
		// b2 needs to be type casted
		manager := b2.(Manager)

		// http packages comes with highly efficient default encoder
		// which makes sure that the data is properly copied onto the response writer
		// We can leverage that implementation or we can write the entire functionality
		// from scratch.

		bt, err := json.Marshal(manager)
		if err != nil {
			return err
		}

		rw.WriteHeader(net_http.StatusOK)
		rw.Write(bt)
		return nil
	}

	// So to stitch all the three above mentioned phases, we will use
	// tr.POST() method which supports a Handler and HandlerOption
	tr.POST(
		"/employee",
		func(cx context.Context, req interface{}) (interface{}, error) {
			return serviceFunc(cx, req.(Employee))
		},
		http.HandlerWithDecoder(decoderFunc),
		http.HandlerWithEncoder(encodingFunc),
	)

	// The above code results in API working as follows
	/*
		➜  ~ curl "http://localhost:4444/employee" -d '{"id":"employee1", "name":"ujjwal"}'
		{
			"id": "employee1-manager",
			"name": "ujjwal-manager"
		}
	*/
	// As you can see the whole construct is based on trying to separate the
	// business logic with rest of the transport.
	// In an event if this library is used for multiple transport, the code
	// in encodingFunc & decodingFunc, can be different based on transport
	// and code in serviceFunc would still remain the same.

	// ErrorHandling
	// By default to handle Error, kit/transport/http.DefaultErrorEncoder is used
	//
	// To handle Custom Error Conditions
	// 	- Write Custom Errors
	// 	- Write Custom ErrorEncoder
	// 	- use http.WithErrorEncoder() to bind Error Encoders
	// To Have custom Error Encoder, use the ServerOption that can be
	// passed using `transport.Get(url, fn, [...ServerOption])` <- these Server Options
	tr.Get("/error", func(
		ctx context.Context,
		req *net_http.Request,
	) (res *net_http.Response, err error) {
		num := rand.Intn(4)
		return nil, errs[num]
	})

	parser := tr.Mux().URLParser()

	// Another Example is of using URL parameters
	tr.Get("/ping/{name}", func(
		ctx context.Context,
		req *net_http.Request,
	) (*net_http.Response, error) {
		params := parser.Parse(req)

		time.Sleep(1 * time.Second)
		return http.NewResponse(
			req,
			http.ResponseWithBytes(
				[]byte("hello "+params.ByName("name")+"!"),
			),
		), nil
	})

	// To start any transport run the transport with Open() method.
	// typically this would run as part of a go-routine, but there is only
	// one tranport in this example, so no go-routine.
	tr.Open()
}

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
