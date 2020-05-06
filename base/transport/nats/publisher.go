package nats

import (
	"context"
	natn "github.com/nats-io/nats.go"
	"github.com/vtomar01/go-base/base/endpoint"
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
		timeout time.Duration
	}
)

// NewPublisher constructs a usable publisher on the same NATS transport.
func NewPublisher(
	con *natn.Conn,
	options ...PublisherOption,
) *Publisher {
	p := &Publisher{
		con: con,
	}
	for _, option := range options {
		option(p)
	}
	return p
}

// Endpoint returns a usable endpoint that invokes the remote endpoint.
func (p *Publisher) Endpoint(subject string, encoder Encoder) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {

		msg := natn.Msg{Subject: subject}

		if err := encoder(ctx, &msg, request); err != nil {
			return nil, err
		}

		err := p.con.PublishMsg(&msg)
		if err != nil {
			return nil, err
		}

		return msg, nil
	}
}
