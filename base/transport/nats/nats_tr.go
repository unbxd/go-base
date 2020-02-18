package nats

import (
	"strings"
	"sync"
	"time"

	"github.com/go-kit/kit/endpoint"
	kitn "github.com/go-kit/kit/transport/nats"
	nats "github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"github.com/unbxd/go-base/base/log"
)

// TransportOption is optional parameters for NATS Transport
type TransportOption func(*Transport)

// WithDisconnectCallback is calback triggered when connection from nats
// server is lost
func WithDisconnectCallback(fn func(nc *nats.Conn, err error)) TransportOption {
	return func(tr *Transport) {
		tr.opts = append(tr.opts, nats.DisconnectErrHandler(fn))
	}
}

// WithReconnectCallback is callback triggered when connection to nats server
// is re-established
func WithReconnectCallback(fn func(nc *nats.Conn)) TransportOption {
	return func(tr *Transport) {
		tr.opts = append(tr.opts, nats.ReconnectHandler(fn))
	}
}

// WithClosedCallback is callback triggered when connection to nats server
// is closed
func WithClosedCallback(fn func(nc *nats.Conn)) TransportOption {
	return func(tr *Transport) {
		tr.opts = append(tr.opts, nats.ClosedHandler(fn))
	}
}

// WithServers lets us customize the list of servers to use with nats
func WithServers(servers []string) TransportOption {
	return func(tr *Transport) {
		tr.srvs = strings.Join(servers, ", ")
	}
}

// WithSeedServers lets you add additional servers which can be used
// as seed servers when you have a nats server running as sidecar
// This creates an experience where the communication from local
// server is used as first priority, however if that fails
// subscriber or publisher here will then try to connect to the
// remainder of seed servers
func WithSeedServers(def string, seeds []string) TransportOption {
	return func(tr *Transport) {
		tr.srvs = def + ", " + strings.Join(seeds, ", ")
		tr.opts = append(tr.opts, nats.DontRandomize())
	}
}

// WithToken allows us to connect to NATS server with Tokens
func WithToken(token string) TransportOption {
	return func(tr *Transport) {
		tr.opts = append(tr.opts, nats.Token(token))
	}
}

// WithDefaultServer sets the application to connect to NATS server
// running in localhost:4222
func WithDefaultServer() TransportOption {
	return func(tr *Transport) {
		tr.srvs = "nats://localhost:4222"
	}
}

// WithFlushTimeout sets a timeout that we will wait for
// publisher to complete flushing the content on nats server
// before terminating connection
func WithFlushTimeout(t time.Duration) TransportOption {
	return func(tr *Transport) {
		tr.flush = t
	}
}

// WithSubscriberOptions sets default subscriber options
func WithSubscriberOptions(opts ...kitn.SubscriberOption) TransportOption {
	return func(tr *Transport) {
		tr.subscribers.defaultOptions = opts
	}
}

// WithLogging sets logging for Transport, subscribers & publishers
func WithLogging(logger log.Logger) TransportOption {
	return func(tr *Transport) {
		if logger == nil {
			return
		}

		tr.logger = logger
	}
}

// Transport is transport server for NATS.IO connection
type Transport struct {
	// nats properties
	srvs string

	conn  *nats.Conn
	opts  []nats.Option
	flush time.Duration

	// internal
	logger log.Logger

	subscribers Subscribers
}

// SubscribeEndpoint provides custom binding for a given event
// This allows for a endpoint to access the message returned
// from the subscriber
func (tr *Transport) SubscribeEndpoint(
	subject string,
	endpoint endpoint.Endpoint,
	decoder kitn.DecodeRequestFunc,
	encoder kitn.EncodeResponseFunc,
	options ...kitn.SubscriberOption,
) (subs *nats.Subscription, err error) {
	ns := kitn.NewSubscriber(
		endpoint,
		decoder,
		encoder,
		options...,
	)

	subs, err = tr.conn.Subscribe(
		subject, ns.ServeMsg(tr.conn),
	)
	return
}

// Open starts the Transport
func (tr *Transport) Open() error {
	return nil
}

// Close shuts down Transport
func (tr *Transport) Close() (err error) {
	// flush and close
	defer tr.conn.Close()

	if tr.flush > 0 {
		err = tr.conn.FlushTimeout(tr.flush)
		if err != nil {
			return
		}
	} else {
		tr.conn.Flush()
	}
	return
}

// NewTransport returns a new nats transport
func NewTransport(
	options ...TransportOption,
) (*Transport, error) {
	var waitgroup sync.WaitGroup

	transport := &Transport{
		conn: nil,
		opts: []nats.Option{},
		subscribers: Subscribers{
			subs:           []Subscriber{},
			defaultOptions: []kitn.SubscriberOption{},
			waitgroup:      &waitgroup,
		},
	}

	for _, o := range options {
		o(transport)
	}

	if transport.conn == nil {
		return nil, errors.New("nats.conn is nil")
	}

	return transport, nil
}
