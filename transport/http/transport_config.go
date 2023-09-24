package http

import (
	"time"

	"github.com/unbxd/go-base/log"
)

type (
	// keyValues is a simple key value
	KeyValue struct {
		Key   string
		Value string
	}

	// config defines the properties used to initialise the
	// transport.
	// this is basically moving the initialisation of the transport
	// to a builder like pattern, where configurations are
	// pre-initialised and then based on configuration, properties
	// are chosen for Transport
	// this method supersedes the old way of creating transport
	// via `NewTransport`
	config struct {
		// server host & port
		host string
		port string

		// time outs for the http.Server
		idle, read, write time.Duration

		// monitor APIs
		monitors []string

		// list of headers which will have requestId
		// other than `x-request-id`
		requestIdHeaders []string

		// logging
		logconfig *logConfig

		// metrics
		metricsconfig *metricsConfig

		// shared handlerOptions by all the APIs
		options []HandlerOption

		// mux can be provided by the application as well
		mux Mux
	}

	metricsConfig struct {
		// enable or disable metrics completely
		enabled bool

		// tags associated with server
		// used by the metrics
		tags []KeyValue

		// enable open telemetry
		openTelemetry bool

		// enable custom metrics via go-base/metrics
		customMetrics bool
	}

	// logConfig is only used if no logger is set in TransportOption
	// or a shortHand functionlike "WithProductionDefaults" or "WithDevelopmentDefaults"
	// are used to create the Transport
	logConfig struct {
		// enable or disable logging completely
		enabled bool

		// properties of logger
		level    string
		encoding string

		// enable tracelogging or not
		traceLogging bool

		// logger
		// used to store logger if provided by the application
		logger log.Logger
	}
)
