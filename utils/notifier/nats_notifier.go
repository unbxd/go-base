package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	natn "github.com/nats-io/nats.go"
	"github.com/pkg/errors"
)

type (
	natsNotifier struct {
		subject string
		options natn.Options
		conn    *natn.Conn
		prefix  string
		name    string
	}

	Option func(*natsNotifier)
)

func (nn *natsNotifier) Notify(
	cx context.Context,
	data interface{},
) error {
	var buff bytes.Buffer

	// serialize it
	err := json.NewEncoder(&buff).Encode(newEvent(nn.name, data))
	if err != nil {
		return errors.Wrap(err, "failed to encode data to json")
	}

	return nn.conn.Publish(
		fmt.Sprintf("%s.%s", nn.prefix, nn.subject),
		buff.Bytes(),
	)
}

func WithDefaultOptions() Option {
	return func(nn *natsNotifier) { nn.options = natn.GetDefaultOptions() }
}

func WithCustomOptions(options natn.Options) Option {
	return func(nn *natsNotifier) { nn.options = options }
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

// NewNotifier returns a default implementation of Notifier, which
// relies on NATS to publish the events.
// Any future implementation should name itself as `New<type>Notifier`
func NewNotifier(
	subject string,
	options ...Option,
) (Notifier, error) {
	nn := &natsNotifier{
		options: natn.GetDefaultOptions(),
		prefix:  "gobase",
		name:    "natsNotifier",
		subject: subject,
	}

	for _, o := range options {
		o(nn)
	}

	cc, err := nn.options.Connect()
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect nats")
	}

	nn.conn = cc

	return nn, nil
}
