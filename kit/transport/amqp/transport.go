package amqp

import (
	"context"
	"errors"
	"sync"
	"time"

	mqp "github.com/streadway/amqp"
	"github.com/unbxd/go-base/utils/log"
)

type (
	// TransportOption is optional parameters for RabbitMQ Transport
	TransportOption func(*Transport)

	Transport struct {
		name string

		server string
		open   bool
		mu     sync.Mutex

		conn *mqp.Connection

		logger      log.Logger
		subscribers map[string]*subscriber

		closeCh chan struct{}
	}

	Subscriber interface {
		Id() string
		Queue() string
		Exchange() string
	}
)

func WithName(name string) TransportOption {
	return func(tr *Transport) {
		tr.name = name
	}
}

func WithServer(server string) TransportOption {
	return func(tr *Transport) {
		tr.server = server
	}
}

// WithLogging sets logging for Transport, subscribers & publishers
func WithLogging(logger log.Logger) TransportOption {
	return func(tr *Transport) {
		tr.logger = logger
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

func (tr *Transport) Publisher(
	options ...PublisherOption,
) (*Publisher, error) {
	return NewPublisher(tr.conn, options...)
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
		tr.conn.Close()
		close(ch)
	}()

	for _, sub := range tr.subscribers {
		_ = sub.close()
	}

	tr.logger.Info("RabbitMQ connection closed..")
	close(tr.closeCh)
	return
}

// NewTransport returns a new RabbitMQ transport
func NewTransport(
	closeCh chan struct{},
	options ...TransportOption,
) (*Transport, error) {
	tr := Transport{
		closeCh:     closeCh,
		subscribers: make(map[string]*subscriber),
	}

	for _, o := range options {
		o(&tr)
	}

	if tr.logger == nil {
		return nil, errors.New("RabbitMQ logger is not set")
	}

	if tr.server == "" {
		return nil, errors.New("RabbitMQ server uri is not set")
	}

	var err error
	tr.conn, err = mqp.Dial(tr.server)
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			_, ok := <-tr.conn.NotifyClose(make(chan *mqp.Error))
			// exit this goroutine if closed by developer
			if !ok {
				break
			}

			// reconnect if not closed by developer
			for {
				time.Sleep(1 * time.Second)

				conn, err := mqp.Dial(tr.server)
				if err == nil {
					tr.conn = conn
					tr.logger.Info("RabbitMQ reconnected successfully")
					break
				}

				tr.logger.Debugf("RabbitMQ reconnect failed, err: %v", err)
			}
		}
	}()

	return &tr, nil
}
