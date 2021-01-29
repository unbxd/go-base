package kafka

import (
	"context"

	"github.com/go-kit/kit/transport"
	"github.com/pkg/errors"
	kafgo "github.com/segmentio/kafka-go"
)

type (

	// BeforeFunc is executed prior to invoking the endpoint. RequestFunc
	// may take information from request recieved in the Consumer
	// and put it in the context.
	// For instance, if the context needs the information about the topic
	// or the group-id, that is populated here
	BeforeFunc func(context.Context, kafgo.Message) context.Context

	// AfterFunc are invoked after executing endpoint
	AfterFunc func(context.Context, kafgo.Message, interface{}) context.Context

	// ErrorFunc handles the error condition
	ErrorFunc func(context.Context, kafgo.Message, error)

	// ErrorHandler is wrapper on top of kit.transport.ErrorHandler
	ErrorHandler interface{ transport.ErrorHandler }
)

// Kafka Errors
var (
	ErrCreatingConsumer = errors.New("error creating consumer")
	ErrCreatingProducer = errors.New("error creating producer")
)
