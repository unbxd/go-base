package nats

import (
	"errors"
)

// NATS Errors
var (
	ErrCreatingSubscriber = errors.New("error creating subscriber")
	ErrCreatingProducer   = errors.New("error creating producer")
)
