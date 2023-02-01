package notifier

import (
	"context"
	"fmt"

	natn "github.com/nats-io/nats.go"

	"github.com/pkg/errors"
	"github.com/unbxd/go-base/kit/transport/nats"
)

type (
	natsNotifier struct {
		*nats.Publisher
		opts    []nats.PublisherOption
		subject string
		prefix  string
		name    string
	}

	Option func(*natsNotifier)
)

func (nn *natsNotifier) Notify(
	cx context.Context,
	data interface{},
) error {
	return nn.Publish(
		cx,
		fmt.Sprintf("%s.%s", nn.prefix, nn.subject),
		data,
	)
}

func WithSubjectPrefix(prefix string) Option {
	return func(nn *natsNotifier) { nn.prefix = prefix }
}

func WithName(name string) Option {
	return func(nn *natsNotifier) { nn.name = name }
}

func WithSubject(subject string) Option {
	return func(nn *natsNotifier) { nn.subject = subject }
}

func WithMessageEncoder(
	fn func(cx context.Context, sub string, data interface{}) (*natn.Msg, error),
) Option {
	return func(p *natsNotifier) {
		p.opts = append(p.opts,
			nats.WithPublishMessageEncoder(fn),
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

	nn := &natsNotifier{
		subject: subject,
	}

	for _, o := range options {
		o(nn)
	}

	pub, err := nats.NewPublisher(connstr, nn.opts...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create publisher")
	}
	nn.Publisher = pub

	return nn, nil
}
