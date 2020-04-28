package nats

import (
	natn "github.com/nats-io/nats.go"
	"github.com/unbxd/go-base/base/endpoint"
	"github.com/unbxd/go-base/base/log"
	"time"
)

type (
	// TransportOption is optional parameters for NATS Transport
	TransportOption func(*Transport)

	// Transport is transport server for natn.IO connection
	Transport struct {
		flush time.Duration

		conn  *natn.Conn
		nopts natn.Options

		logger      log.Logger
		subscribers []subscriber

		closeCh chan struct{}
	}
)

func WithDisconnectCallback(fn func(nc *natn.Conn, err error)) TransportOption {
	return func(tr *Transport) {
		tr.nopts.DisconnectedErrCB = fn
	}
}

func WithReconnectCallback(fn func(nc *natn.Conn)) TransportOption {
	return func(tr *Transport) {
		tr.nopts.ReconnectedCB = fn
	}
}

func WithClosedCallback(fn func(nc *natn.Conn)) TransportOption {
	return func(tr *Transport) {
		tr.nopts.ClosedCB = fn
	}
}

func WithServers(servers []string) TransportOption {
	return func(tr *Transport) {
		tr.nopts.Servers = servers
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

// WithLogging sets logging for Transport, subscribers & publishers
func WithLogging(logger log.Logger) TransportOption {
	return func(tr *Transport) {
		tr.logger = logger
	}
}

func (tr *Transport) Subscribe(
	options ...SubscriberOption,
) error {
	s, err := newSubscriber(tr.logger, options...)
	if err != nil {
		return err
	}
	tr.subscribers = append(tr.subscribers, *s)
	return nil
}

func (tr *Transport) Publisher(
	subject string,
	enc Encoder,
	options ...PublisherOption,
) endpoint.Endpoint {
	return newPublisher(tr.conn, subject, enc, options...).endpoint()
}

// Open starts the Transport
func (tr *Transport) Open() error {

	var err error

	tr.conn, err = tr.nopts.Connect()
	if err != nil {
		return err
	}

	for _, sub := range tr.subscribers {
		err = sub.open(tr.conn)
		if err != nil {
			return err
		}
	}
	return nil
}

// Close shuts down Transport
func (tr *Transport) Close() (err error) {

	// flush and close
	defer tr.conn.Close()

	for _, sub := range tr.subscribers {
		_ = sub.close()
	}

	if tr.flush > 0 {
		err = tr.conn.FlushTimeout(tr.flush)
		if err != nil {
			return
		}
	} else {
		_ = tr.conn.Flush()
	}
	return
}

func (tr *Transport) onClose(_ *natn.Conn) {
	close(tr.closeCh)
}

// NewTransport returns a new NATS transport
func NewTransport(
	options ...TransportOption,
) (*Transport, error) {

	tr := Transport{nopts: natn.GetDefaultOptions()}

	for _, o := range options {
		o(&tr)
	}

	tr.closeCh = make(chan struct{})
	tr.nopts.ClosedCB = tr.onClose

	return &tr, nil
}
