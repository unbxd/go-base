package amqp

import (
	"errors"
)

// RabbitMQ Errors
var (
	ErrCreatingSubscriber = errors.New("error creating subscriber")
	ErrCreatingPublisher  = errors.New("error creating publisher")
	ErrCreatingExchange   = errors.New("error creating exchange")
	ErrCreatingQueue      = errors.New("error creating queue")
)
