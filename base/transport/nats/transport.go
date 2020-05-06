package nats

import (
	"context"
	"errors"
	natn "github.com/nats-io/nats.go"
	"github.com/vtomar01/go-base/base/log"
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

func WithServers(servers []string) TransportOption {
	return func(tr *Transport) {
		tr.nopts.Servers = servers
	}
}

func WithNoRandomize(noRandomize bool) TransportOption {
	return func(tr *Transport) {
		tr.nopts.NoRandomize = noRandomize
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

func WithName(n string) TransportOption {
	return func(tr *Transport) {
		tr.nopts.Name = n
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
	options ...PublisherOption,
) *Publisher {
	return NewPublisher(tr.conn, options...)
}

// Open starts the Transport
func (tr *Transport) Open() error {

	for _, sub := range tr.subscribers {
		err := sub.open(tr.conn)
		if err != nil {
			return err
		}
	}
	return nil
}

// Close shuts down Transport
func (tr *Transport) Close() (err error) {

	ctx, cancel := context.WithTimeout(
		context.Background(), 100*time.Second,
	)
	defer cancel()

	ch := make(chan struct{})
	go func() {
		err = tr.close(ch)

	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ch:
			return
		}
	}
}

func (tr *Transport) close(ch chan struct{}) (err error) {
	defer func() {
		// flush and close
		tr.conn.Close()
		close(ch)
	}()

	for _, sub := range tr.subscribers {
		_ = sub.close()
	}

	if tr.flush > 0 {
		err = tr.conn.FlushTimeout(tr.flush)
	} else {
		_ = tr.conn.Flush()
	}
	return
}

func (tr *Transport) onClose(_ *natn.Conn) {
	tr.logger.Info("NATS connection closed..")
	close(tr.closeCh)
}

// NewTransport returns a new NATS transport
func NewTransport(
	closeCh chan struct{},
	options ...TransportOption,
) (*Transport, error) {

	tr := Transport{
		nopts:   natn.GetDefaultOptions(),
		closeCh: closeCh,
	}

	for _, o := range options {
		o(&tr)
	}

	if tr.logger == nil {
		return nil, errors.New("NATS logger is not set")
	}

	tr.nopts.ClosedCB = tr.onClose

	var err error
	tr.conn, err = tr.nopts.Connect()
	if err != nil {
		return nil, err
	}

	return &tr, nil
}
