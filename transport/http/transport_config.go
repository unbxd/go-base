package http

import (
	"net/http"
	"time"

	kit_http "github.com/go-kit/kit/transport/http"
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
		// server name
		name    string
		version string

		// server host & port
		host string
		port string

		heartbeats []string

		// time outs for the http.Server
		idleTimeout, readTimeout, writeTimeout time.Duration

		logging      bool
		traceLogging bool
		logger       log.Logger

		// metrics
		metrics bool

		// shared handlerOptions by all the APIs
		transportOptions []TransportOption
		handlerOptions   []HandlerOption

		// transport level ffs, which applies to all paths
		ffs []Filter

		// mux can be provided by the application as well
		// default is nil, which means default multiplexer
		// is used
		muxOptions []DefaultMuxOption

		panicFormatter PanicFormatter
	}

	TransportConfigOption func(*config) error
)

func (c *config) filters() []Filter {
	// default filters available by default to all routes
	filters := []Filter{
		closerFilter(), // closes the request
		panicRecoveryFilter( // handles panic
			c.logger,
			withCustomFormatter(c.panicFormatter),
			withStack(1024*8, false),
		),
		wrappedResponseWriterFilter(), // wraps response for easy status access
		serverNameFilter(c.name, c.version),
		requestIDFilter(),
		decorateContextFilter(),
		heartbeatFilter(c.name, c.heartbeats), // heartbeats for filter

	}

	if c.logging && c.traceLogging {
		filters = append(filters, TraceLoggingFilter(c.logger))
	}

	// append rest of our filters
	filters = append(filters, c.ffs...)
	return filters
}

func (c *config) build() (*Transport, error) {
	tr := &Transport{
		Server:  &http.Server{Addr: c.host + ":" + c.port},
		name:    c.name,
		logger:  c.logger,
		muxer:   NewDefaultMux(c.muxOptions...),
		options: c.handlerOptions,
	}

	for _, fn := range c.transportOptions {
		fn(tr)
	}

	tr.Handler = Chain(tr.muxer, c.filters()...)
	return tr, nil
}

func newConfig(name string) *config {

	logger, _ := log.NewZeroLogger(
		log.ZeroLoggerWithAsyncSink(1000, 2, nil),
		log.ZeroLoggerWithFields(log.String("server", name)),
		log.ZeroLoggerWithLevel("error"),
	)

	return &config{
		name:             name,
		version:          "v0.0.0",
		host:             "0.0.0.0",
		port:             "7001",
		heartbeats:       []string{"/ping"},
		idleTimeout:      90 * time.Second,
		readTimeout:      5 * time.Second,
		writeTimeout:     10 * time.Second,
		logging:          true,
		traceLogging:     true,
		logger:           logger,
		metrics:          true,
		transportOptions: []TransportOption{},
		handlerOptions: []HandlerOption{
			NewErrorEncoderHandlerOptions(kit_http.DefaultErrorEncoder),
		},
		ffs:            []Filter{},
		muxOptions:     []DefaultMuxOption{},
		panicFormatter: &textPanicFormatter{},
	}
}

func NewHTTPTransport(
	name string, options ...TransportConfigOption,
) (*Transport, error) {
	cfg := newConfig(name)

	for _, ofn := range options {
		ofn(cfg)
	}

	return cfg.build()
}
