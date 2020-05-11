package nats

import (
	"context"
	natn "github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"github.com/unbxd/go-base/base/endpoint"
	"time"
)

type (
	// PublisherOption lets you modify properties for publisher
	PublisherOption func(*Publisher)

	// Encoder encodes the value passed to it and converts to NATS message
	Encoder func(context.Context, *natn.Msg, interface{}) error

	// publisher publishes message on NATS
	Publisher struct {
		con     *natn.Conn
		encoder Encoder
		timeout time.Duration
	}
)

func WithEncoder(e Encoder) PublisherOption {
	return func(p *Publisher) {
		p.encoder = e
	}
}

// NewPublisher constructs a usable publisher on the same NATS transport.
func NewPublisher(
	con *natn.Conn,
	options ...PublisherOption,
) (*Publisher, error) {
	p := &Publisher{
		con: con,
	}
	for _, option := range options {
		option(p)
	}

	if p.encoder == nil {
		return nil, errors.Wrap(ErrCreatingPublisher, "encoder is nil")
	}
	return p, nil
}

// Endpoint returns a usable endpoint that invokes the remote endpoint.
func (p *Publisher) Endpoint(subject string) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		return p.publish(ctx, subject, request)
	}
}

//publish to given topic
func (p *Publisher) Publish(
	ctx context.Context,
	subject string,
	request interface{},
) (interface{}, error) {

	return p.publish(ctx, subject, request)
}

func (p *Publisher) publish(
	ctx context.Context,
	subject string,
	request interface{},
) (interface{}, error) {
	msg := natn.Msg{Subject: subject}

	if err := p.encoder(ctx, &msg, request); err != nil {
		return nil, err
	}

	err := p.con.PublishMsg(&msg)
	if err != nil {
		return nil, err
	}

	return msg, nil
}
