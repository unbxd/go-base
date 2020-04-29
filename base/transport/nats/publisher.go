package nats

import (
	"context"
	natn "github.com/nats-io/nats.go"
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
		timeout time.Duration
	}
)

// NewPublisher constructs a usable publisher on the same NATS transport.
func NewPublisher(
	con *natn.Conn,
	options ...PublisherOption,
) *Publisher {
	p := &Publisher{
		con:     con,
		timeout: 10 * time.Second,
	}
	for _, option := range options {
		option(p)
	}
	return p
}

// PublisherTimeout sets the available timeout for NATS request.
func PublisherTimeout(timeout time.Duration) PublisherOption {
	return func(p *Publisher) { p.timeout = timeout }
}

// Endpoint returns a usable endpoint that invokes the remote endpoint.
func (p *Publisher) Endpoint(subject string, encoder Encoder) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		ctx, cancel := context.WithTimeout(ctx, p.timeout)
		defer cancel()

		msg := natn.Msg{Subject: subject}

		if err := encoder(ctx, &msg, request); err != nil {
			return nil, err
		}

		err := p.con.Publish(msg.Subject, msg.Data)
		if err != nil {
			return nil, err
		}

		return msg, nil
	}
}
