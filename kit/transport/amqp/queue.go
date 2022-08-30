package amqp

import (
	"github.com/pkg/errors"
	mqp "github.com/streadway/amqp"
)

type (
	queue struct {
		name       string
		durable    bool
		autoDelete bool
		exclusive  bool
		noWait     bool
		args       mqp.Table
	}

	// QueueOption provides set of options to modify a Queue
	QueueOption func(*queue)
)

func NewQueue(name string, opts ...QueueOption) (*queue, error) {

	if name == "" {
		return nil, errors.Wrap(ErrCreatingQueue, "queue name is empty")
	}

	// setting default values for queue
	q := &queue{
		name:       name,
		durable:    false,
		autoDelete: false,
		exclusive:  false,
		noWait:     false,
		args:       nil,
	}

	// default settings can be overwritten by options
	for _, option := range opts {
		option(q)
	}

	return q, nil
}

func WithDurableQueueOption(durable bool) QueueOption {
	return func(q *queue) {
		q.durable = durable
	}
}

func WithAutoDeleteQueueOption(autoDelete bool) QueueOption {
	return func(q *queue) {
		q.autoDelete = autoDelete
	}
}

func WithExclusiveQueueOption(exclusive bool) QueueOption {
	return func(q *queue) {
		q.exclusive = exclusive
	}
}

func WithNoWaitQueueOption(noWait bool) QueueOption {
	return func(q *queue) {
		q.noWait = noWait
	}
}

func WithArgsQueueOption(args mqp.Table) QueueOption {
	return func(q *queue) {
		q.args = args
	}
}
