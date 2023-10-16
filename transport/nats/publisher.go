package nats

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	natn "github.com/nats-io/nats.go"
	"github.com/unbxd/go-base/v2/endpoint"
	"github.com/unbxd/go-base/v2/errors"
)

type (
	// PublisherOption lets you modify properties for publisher
	PublisherOption func(*Publisher)

	// PublishMessageEncoder encodes the value passed to it and converts to NATS message
	PublishMessageEncoder func(cx context.Context, sub string, data interface{}) (*natn.Msg, error)

	// Before is a function that is called before every message sent to NATS
	BeforePublish func(context.Context, *natn.Msg) error

	// After is a function called after every message sent to NATS
	AfterPublish func(context.Context, *natn.Msg, error)

	// PublishErrorHandler is a function that is called when an error occurs
	PublishErrorHandler func(context.Context, error) error

	// publisher publishes message on NATS
	Publisher struct {
		conn *natn.Conn
		opts *natn.Options

		name   string
		prefix string

		encoder      PublishMessageEncoder
		befores      []BeforePublish
		afters       []AfterPublish
		errorHandler PublishErrorHandler

		headers natn.Header
	}
)

func WithPublishMessageEncoder(encoder PublishMessageEncoder) PublisherOption {
	return func(p *Publisher) {
		p.encoder = encoder
	}
}

func WithBeforePublish(befores ...BeforePublish) PublisherOption {
	return func(p *Publisher) {
		p.befores = append(p.befores, befores...)
	}
}

func WithAfterPublish(afters ...AfterPublish) PublisherOption {
	return func(p *Publisher) {
		p.afters = append(p.afters, afters...)
	}
}

func WithErrorHandler(handler PublishErrorHandler) PublisherOption {
	return func(p *Publisher) {
		p.errorHandler = handler
	}
}

func WithPublisherName(name string) PublisherOption {
	return func(p *Publisher) {
		p.name = name
	}
}

func WithPublisherSubjectPrefix(prefix string) PublisherOption {
	return func(p *Publisher) {
		p.prefix = prefix
	}
}

func WithCustomPublisherTimeout(timeout time.Duration) PublisherOption {
	return func(p *Publisher) {
		p.opts.Timeout = timeout
	}
}

func WithCustomPublisherMaxReconnect(maxReconnect int) PublisherOption {
	return func(p *Publisher) {
		p.opts.MaxReconnect = maxReconnect
	}
}

func WithCustomPublisherPingInterval(pingInterval time.Duration) PublisherOption {
	return func(p *Publisher) {
		p.opts.PingInterval = pingInterval
	}
}

func WithPublishHeader(headers natn.Header) PublisherOption {
	return func(p *Publisher) {
		p.headers = headers
	}
}

func defaultPublishMessageEncoder(
	cx context.Context, subject string, data interface{},
) (*natn.Msg, error) {
	var (
		buf bytes.Buffer
		err error
	)

	err = json.NewEncoder(&buf).Encode(data)
	if err != nil {
		return nil, errors.Wrap(err, "defaultencoder: encoding error")
	}

	return &natn.Msg{
		Subject: subject,
		Data:    buf.Bytes(),
	}, err
}

func defaultPublishErrorHandler(cx context.Context, err error) error {
	return err
}

func NewPublisher(connstr string, options ...PublisherOption) (*Publisher, error) {
	var (
		err  error
		cc   *natn.Conn
		opts = natn.GetDefaultOptions()
		pb   = &Publisher{
			conn:         nil,
			opts:         &opts,
			name:         "go-base-publisher",
			prefix:       "gb",
			encoder:      defaultPublishMessageEncoder,
			befores:      []BeforePublish{},
			afters:       []AfterPublish{},
			errorHandler: defaultPublishErrorHandler,
		}
	)

	for _, fn := range options {
		fn(pb)
	}

	pb.opts.Url = connstr

	cc, err = pb.opts.Connect()
	if err != nil {
		return nil, errors.Wrap(
			err, "unable to connect to nats server",
		)
	}

	pb.conn = cc
	return pb, err
}

func subject(prefix, subject string) string {
	if prefix != "" {
		return fmt.Sprintf("%s.%s", prefix, subject)
	}
	return subject
}

// Endpoint returns a usable endpoint that invokes the remote endpoint.
func (p *Publisher) Endpoint(sub string) endpoint.Endpoint {
	return func(ctx context.Context, data interface{}) (interface{}, error) {
		return p.send(ctx, subject(p.prefix, sub), data)
	}
}

// Publish publishes the message on NATS
func (p *Publisher) Publish(ctx context.Context, sub string, data interface{}) error {
	_, err := p.send(ctx, subject(p.prefix, sub), data)
	return err
}

func (p *Publisher) send(cx context.Context, sub string, data interface{}) (*natn.Msg, error) {
	msg, err := p.encoder(cx, sub, data)
	if err != nil {
		return nil, p.errorHandler(cx, err)
	}

	for _, fn := range p.befores {
		err := fn(cx, msg)
		if err != nil {
			return nil, p.errorHandler(cx, err)
		}
	}

	err = p.conn.PublishMsg(msg)
	if err != nil {
		return nil, p.errorHandler(cx, err)
	}

	for _, fn := range p.afters {
		fn(cx, msg, err)
	}

	return msg, err
}
