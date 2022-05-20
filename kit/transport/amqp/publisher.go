package amqp

import (
	"context"
	"time"

	kita "github.com/go-kit/kit/transport/amqp"
	"github.com/pkg/errors"
	mqp "github.com/streadway/amqp"
	"github.com/unbxd/go-base/kit/endpoint"
)

type (
	// PublisherOption lets you modify properties for publisher
	PublisherOption func(*Publisher)

	// Encoder encodes the value passed to it and converts to AMQP message
	Encoder func(context.Context, *mqp.Publishing, interface{}) error

	Publisher struct {
		channel   *mqp.Channel
		q         *mqp.Queue
		exchange  string
		deliverer kita.Deliverer
		encoder   Encoder
		timeout   time.Duration
	}
)

func WithEncoder(e Encoder) PublisherOption {
	return func(p *Publisher) {
		p.encoder = e
	}
}

func WithDeliverer(d kita.Deliverer) PublisherOption {
	return func(p *Publisher) {
		p.deliverer = d
	}
}

func WithQueue(q *mqp.Queue) PublisherOption {
	return func(p *Publisher) {
		p.q = q
	}
}

func WithExchange(exchange string) PublisherOption {
	return func(p *Publisher) {
		p.exchange = exchange
	}
}

// NewPublisher constructs a usable publisher on the same RabbitMQ transport.
func NewPublisher(
	conn *mqp.Connection,
	options ...PublisherOption,
) (*Publisher, error) {
	p := &Publisher{}
	for _, option := range options {
		option(p)
	}

	if p.encoder == nil {
		return nil, errors.Wrap(ErrCreatingPublisher, "encoder is nil")
	}

	if p.exchange == "" {
		return nil, errors.Wrap(ErrCreatingPublisher, "exchange is empty")
	}

	if p.deliverer == nil {
		p.deliverer = kita.SendAndForgetDeliverer
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, errors.Wrap(ErrCreatingPublisher, "creating channel failed")
	}
	p.channel = ch

	err = p.channel.ExchangeDeclare(
		p.exchange,
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, errors.Wrap(ErrCreatingPublisher, "declaring exchange failed")
	}

	return p, nil
}

// Endpoint returns a usable endpoint that invokes the remote endpoint.
func (p *Publisher) Endpoint(exchange string) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		return p.publish(ctx, exchange, request)
	}
}

//publish to given exchange
func (p *Publisher) Publish(
	ctx context.Context,
	exchange string,
	request interface{},
) error {

	_, err := p.publish(ctx, exchange, request)
	return err
}

func (p *Publisher) publish(
	ctx context.Context,
	exchange string,
	request interface{},
) (interface{}, error) {
	msg := mqp.Publishing{}

	if err := p.encoder(ctx, &msg, request); err != nil {
		return nil, err
	}

	err := p.channel.Publish(exchange, "", false, false, msg)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
