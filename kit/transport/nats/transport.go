package nats

import (
	"context"
	"errors"
	"sync"
	"time"

	natn "github.com/nats-io/nats.go"
	"github.com/unbxd/go-base/utils/log"
)

type (
	// TransportOption is optional parameters for NATS Transport
	TransportOption func(*Transport)

	ConnectionErrHandler func(t *Transport, e error)
	// Transport is transport server for natn.IO connection
	Transport struct {
		open  bool
		mu    sync.Mutex
		flush time.Duration

		conn  *natn.Conn
		nopts natn.Options

		logger      log.Logger
		subscribers map[string]*subscriber

		closeCh chan struct{}
	}

	Subscriber interface {
		Id() string
		Topic() string
		Group() string
		IsValid() bool
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

// WithConnectionErrorHandler sets a handler for connection errors
func WithConnectionErrorHandler(h ConnectionErrHandler) TransportOption {
	return func(t *Transport) {
		t.conn.SetErrorHandler(
			func(c *natn.Conn, s *natn.Subscription, e error) {
				h(t, e)
			},
		)
	}
}

func (tr *Transport) Subscribers() []Subscriber {
	var ss []Subscriber
	for _, s := range tr.subscribers {
		ss = append(ss, s)
	}
	return ss
}

func (tr *Transport) Subscribe(
	options ...SubscriberOption,
) (Subscriber, error) {

	s, err := newSubscriber(tr.logger, tr.conn, options...)
	if err != nil {
		return nil, err
	}

	if tr.open {
		err := s.open()
		if err != nil {
			return nil, err
		}
	}

	tr.subscribers[s.id] = s
	return s, nil
}

func (tr *Transport) Unsubscribe(id string) error {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	s, ok := tr.subscribers[id]
	if !ok {
		return nil
	}
	err := s.close()
	if err != nil {
		return err
	}
	delete(tr.subscribers, id)
	return nil
}

// Open starts the Transport
func (tr *Transport) Open() error {

	for _, sub := range tr.subscribers {
		err := sub.open()
		if err != nil {
			return err
		}
	}
	tr.open = true
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
		nopts:       natn.GetDefaultOptions(),
		closeCh:     closeCh,
		subscribers: make(map[string]*subscriber),
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
