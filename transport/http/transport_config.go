package http

import (
	"net/http"
	"time"

	kit_http "github.com/go-kit/kit/transport/http"
	"github.com/unbxd/go-base/v2/log"
)

type (
	// keyValues is a simple key value
	KeyValue struct {
		Key   string
		Value string
	}

	keyValues []KeyValue

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

		logger log.Logger

		// shared handlerOptions by all the APIs
		transportOptions []TransportOption

		// transport level ffs, which applies to all paths
		ffs []Filter

		// mux can be provided by the application as well
		// default is nil, which means default multiplexer
		// is used
		muxOptions []ChiMuxOption

		panicFormatter PanicFormatter
	}

	TransportConfigOption func(*config) error
)

func (kv *KeyValue) String() string    { return kv.Key + ":" + kv.Value }
func (kv *KeyValue) Strings() []string { return []string{kv.Key, kv.Value} }

func (kvs keyValues) tags() []string {
	ts := make([]string, 0)

	for _, kv := range kvs {
		ts = append(ts, kv.Strings()...)
	}

	return ts
}

func (c *config) filters() []Filter {
	// default filters available by default to all routes
	filters := []Filter{
		noopFilter(),
		panicRecoveryFilter( // handles panic
			c.logger,
			WithCustomFormatter(c.panicFormatter),
			WithStack(1024*8, false),
		),
		heartbeatFilter(c.name, c.heartbeats), // heartbeats for filter
		serverNameFilter(c.name, c.version),
		wrappedResponseWriterFilter(), // wraps response for easy status access
		decorateContextFilter(),
		requestIDFilter(),
	}
	return filters
}

func (c *config) build() (*Transport, error) {
	tr := &Transport{
		Server: &http.Server{
			Addr:         c.host + ":" + c.port,
			IdleTimeout:  c.idleTimeout,
			ReadTimeout:  c.readTimeout,
			WriteTimeout: c.writeTimeout,
		},

		name:           c.name,
		logger:         c.logger,
		muxer:          newChiMux(c.muxOptions...),
		handlerOptions: []HandlerOption{},
	}

	for _, fn := range c.transportOptions {
		fn(tr)
	}

	tr.muxer.Use(c.ffs...)

	tr.Handler = chain(tr.muxer, c.filters()...)

	return tr, nil
}

func newConfig(name string) *config {

	logger, _ := log.NewZeroLogger(
		log.ZeroLoggerWithAsyncSink(1000, 2, nil),
		log.ZeroLoggerWithFields(log.String("server", name)),
		log.ZeroLoggerWithLevel("error"),
	)

	return &config{
		name:         name,
		version:      "v0.0.0",
		host:         "0.0.0.0",
		port:         "7001",
		heartbeats:   []string{"/ping"},
		idleTimeout:  90 * time.Second,
		readTimeout:  5 * time.Second,
		writeTimeout: 10 * time.Second,
		logger:       logger,
		transportOptions: []TransportOption{
			WithHandlerOption(
				NewErrorEncoderHandlerOptions(kit_http.DefaultErrorEncoder),
			),
		},
		ffs:            []Filter{},
		muxOptions:     []ChiMuxOption{},
		panicFormatter: &textPanicFormatter{},
	}
}

func NewHTTPTransport(
	name string,
	options ...TransportConfigOption,
) (*Transport, error) {
	cfg := newConfig(name)

	for _, ofn := range options {
		er := ofn(cfg)

		if er != nil {
			return nil, er
		}
	}

	return cfg.build()
}
