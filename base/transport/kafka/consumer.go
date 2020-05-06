package kafka

import (
	"context"

	"time"

	"github.com/go-kit/kit/transport"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	kafgo "github.com/segmentio/kafka-go"
	"github.com/vtomar01/go-base/base/endpoint"
	"github.com/vtomar01/go-base/base/log"
)

type (
	// ConsumerOption provies set of options to modify a subscriber
	ConsumerOption func(*Consumer)

	// Decoder decodes the message recieved on Kafka and converts in
	// business logic
	Decoder func(context.Context, kafgo.Message) (interface{}, error)

	// Consumer is kafka Consumer
	Consumer struct {
		autocommit bool

		reader *kafgo.Reader
		config *kafgo.ReaderConfig

		end     endpoint.Endpoint
		dec     Decoder
		befores []BeforeFunc
		afters  []AfterFunc
		errFn   ErrorFunc

		errHandler ErrorHandler
	}
)

const (
	defaultConsumerGroupID = "go-base-consumer"
	defaultTopic           = "go-base-test"
)

// WithGroupIDConsumerOption provides an option to modify the GroupID for
// a consumer Group
func WithGroupIDConsumerOption(groupID string) ConsumerOption {
	return func(c *Consumer) {
		c.config.GroupID = groupID
	}
}

// WithTopicConsumerOption provides an option to modify the topic
// on which the Consumer will listen to
func WithTopicConsumerOption(topic string) ConsumerOption {
	return func(c *Consumer) {
		c.config.Topic = topic
	}
}

// WithMaxMinByteConsumerOption provi-des an option to modify the min/max
// byte that can written to kafka
func WithMaxMinByteConsumerOption(min, max int) ConsumerOption {
	return func(c *Consumer) {
		c.config.MinBytes = min
		c.config.MaxBytes = max
	}
}

// WithAutoCommitConsumerOption sets the autocommit property of consumer
func WithAutoCommitConsumerOption(flag bool) ConsumerOption {
	return func(c *Consumer) { c.autocommit = flag }
}

// WithAutoCommitTimeConsumerOption sets the auto commit time for Consumer
func WithAutoCommitTimeConsumerOption(dur time.Duration) ConsumerOption {
	return func(c *Consumer) { c.config.CommitInterval = dur }
}

// WithDecoderConsumerOption sets the decoder for the Consumer Message
func WithDecoderConsumerOption(fn Decoder) ConsumerOption {
	return func(c *Consumer) { c.dec = fn }
}

// WithErrorFuncConsumerOption provides a callback to handle error
func WithErrorFuncConsumerOption(fn ErrorFunc) ConsumerOption {
	return func(c *Consumer) { c.errFn = fn }
}

// WithBeforeFuncsConsumerOption provides a way to set BeforeFunc(s)
// to the consumer
func WithBeforeFuncsConsumerOption(fns ...BeforeFunc) ConsumerOption {
	return func(c *Consumer) { c.befores = append(c.befores, fns...) }
}

// WithAfterFuncsConsumerOption provides a way to set AfterFunc(s)
// to the consumer
func WithAfterFuncsConsumerOption(fns ...AfterFunc) ConsumerOption {
	return func(c *Consumer) { c.afters = append(c.afters, fns...) }
}

// WithEndpointConsumerOption provides a way to set
// endpoint to the consumer
func WithEndpointConsumerOption(end endpoint.Endpoint) ConsumerOption {
	return func(c *Consumer) { c.end = end }
}

// WithReaderConsumerOption lets you set the reader for kafka
func WithReaderConsumerOption(reader *kafgo.Reader) ConsumerOption {
	return func(c *Consumer) { c.reader = reader }
}

// Open actually handles the subcriber messages
func (c *Consumer) Open() error {
	if c.reader == nil {
		c.reader = kafgo.NewReader(*c.config)
	}

	for {
		// start a new context
		var (
			ctx = context.Background()
			msg kafgo.Message
			err error
		)

		if c.autocommit {
			msg, err = c.reader.ReadMessage(ctx)
		} else {
			msg, err = c.reader.FetchMessage(ctx)
		}

		if err != nil {
			c.errFn(ctx, msg, errors.Wrap(
				err, "read message from kafka failed",
			))
			c.errHandler.Handle(ctx, err)
			continue
		}

		// before endpoint
		for _, fn := range c.befores {
			ctx = fn(ctx, msg)
		}

		rq, err := c.dec(ctx, msg)
		if err != nil {
			c.errFn(ctx, msg, err)
			c.errHandler.Handle(ctx, err)
			continue
		}

		// execute endpoint
		rs, err := c.end(ctx, rq)
		if err != nil {
			c.errFn(ctx, msg, err)
			c.errHandler.Handle(ctx, err)
			continue
		}

		for _, fn := range c.afters {
			ctx = fn(ctx, msg, rs)
		}

		if !c.autocommit {
			err = c.reader.CommitMessages(ctx, msg)
			if err != nil {
				c.errFn(ctx, msg, err)
				c.errHandler.Handle(ctx, err)
				continue
			}
		}
	}
}

// NewConsumer returns kafka consumer for the given brokers
func NewConsumer(
	brokers []string,
	logger log.Logger,
	options ...ConsumerOption,
) (*Consumer, error) {
	// default values
	cfg := kafgo.ReaderConfig{
		Brokers: brokers,
		GroupID: defaultConsumerGroupID,
		Topic:   defaultTopic,
		Logger:  kafka.LoggerFunc(logger.Debugf),
	}

	cs := &Consumer{
		reader: nil,
		config: &cfg,
	}

	for _, o := range options {
		o(cs)
	}

	if cs.end == nil {
		return nil, errors.Wrap(
			ErrCreatingConsumer, "missing endpoint",
		)
	}

	if cs.dec == nil {
		return nil, errors.Wrap(
			ErrCreatingConsumer, "missing decoder",
		)
	}

	if cs.errFn == nil {
		cs.errFn = defaultErrorFunc
	}

	if cs.errHandler == nil {
		cs.errHandler = transport.NewLogErrorHandler(logger)
	}
	return cs, nil
}
