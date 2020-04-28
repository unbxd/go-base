package nats

import (
	"errors"
)

// NATS Errors
var (
	ErrCreatingSubscriber = errors.New("error creating subscriber")
	ErrCreatingPublisher  = errors.New("error creating publisher")
)
