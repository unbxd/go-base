package notifier

import (
	"context"

	"github.com/pkg/errors"
	"github.com/unbxd/go-base/kit/transport/nats"
)

type (
	writer interface {
		Write(cx context.Context, subject string, data interface{}) error
	}

	defaultWriter struct{ *natsNotifier }
)

func (dw *defaultWriter) Write(
	cx context.Context,
	sub string,
	data interface{},
) error {
	return dw.Publish(cx, sub, data)
}

func newDefaultWriter(nn *natsNotifier) writer { return &defaultWriter{nn} }

type (
	Option  func(*natsNotifier)
	Encoder nats.PublishMessageEncoder

	natsNotifier struct {
		*nats.Publisher

		writer writer

		opts    []nats.PublisherOption
		subject string
	}
)

func (nn *natsNotifier) Notify(
	cx context.Context, data interface{},
) error {
	return nn.writer.Write(
		cx,
		nn.subject,
		data,
	)
}

// Options

func WithSubjectPrefix(prefix string) Option {
	return func(nn *natsNotifier) {
		nn.opts = append(
			nn.opts,
			nats.WithPublisherSubjectPrefix(prefix),
		)
	}
}

func WithMessageEncoder(fn Encoder) Option {
	return func(p *natsNotifier) {
		p.opts = append(
			p.opts,
			nats.WithPublishMessageEncoder(
				nats.PublishMessageEncoder(fn),
			),
		)
	}
}

// NewNotifier returns a default implementation of Notifier, which
// relies on NATS to publish the events.
// Any future implementation should name itself as `New<type>Notifier`
func NewNotifier(
	connstr string,
	subject string,
	options ...Option,
) (Notifier, error) {
	var nn *natsNotifier

	nn = &natsNotifier{
		subject: subject,
		writer:  newDefaultWriter(nn),
	}

	for _, fn := range options {
		fn(nn)
	}

	pub, err := nats.NewPublisher(
		connstr, nn.opts...,
	)

	if err != nil {
		return nil, errors.Wrap(
			err,
			"failed to create publisher",
		)
	}

	nn.Publisher = pub
	return nn, nil
}
