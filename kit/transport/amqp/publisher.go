package amqp

import (
	"context"
	"fmt"

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
		key       string
		exchange  *exchange
		deliverer kita.Deliverer
		encoder   Encoder
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

func WithRoutingKey(key string) PublisherOption {
	return func(p *Publisher) {
		p.key = key
	}
}

func WithExchange(e *exchange) PublisherOption {
	return func(p *Publisher) {
		p.exchange = e
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

	if p.exchange == nil {
		return nil, errors.Wrap(ErrCreatingPublisher, "exchange is nil")
	}

	if p.key == "" {
		return nil, errors.Wrap(ErrCreatingPublisher, "routing key is empty")
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
		p.exchange.name,
		p.exchange.kind,
		p.exchange.durable,
		p.exchange.autoDelete,
		p.exchange.internal,
		p.exchange.noWait,
		p.exchange.args,
	)
	if err != nil {
		fmt.Println("err", err.Error())
		return nil, errors.Wrap(ErrCreatingPublisher, "declaring exchange failed")
	}

	return p, nil
}

// Endpoint returns a usable endpoint that invokes the remote endpoint.
func (p *Publisher) Endpoint() endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		return nil, p.publish(ctx, request)
	}
}

//publish to given exchange
func (p *Publisher) Publish(
	ctx context.Context,
	request interface{},
) error {
	return p.publish(ctx, request)
}

func (p *Publisher) publish(
	ctx context.Context,
	request interface{},
) error {
	msg := mqp.Publishing{}

	if err := p.encoder(ctx, &msg, request); err != nil {
		return err
	}

	err := p.channel.Publish(p.exchange.name, p.key, false, false, msg)
	if err != nil {
		return err
	}

	return nil
}
