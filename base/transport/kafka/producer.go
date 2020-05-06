package kafka

import (
	"context"

	"github.com/go-kit/kit/transport"
	"github.com/pkg/errors"
	kafgo "github.com/segmentio/kafka-go"
	"github.com/vtomar01/go-base/base/endpoint"
	"github.com/vtomar01/go-base/base/log"
)

type (
	// ProducerOption lets you modify properties for
	// producer
	ProducerOption func(*Producer)

	// Encoder encodes the value passed to it and converts to
	// kafka message
	Encoder func(context.Context, interface{}) (kafgo.Message, error)

	// Producer produces message on Kafka
	Producer struct {
		writer *kafgo.Writer
		config *kafgo.WriterConfig

		enc Encoder

		befores []BeforeFunc
		afters  []AfterFunc
		errFn   ErrorFunc

		errHn ErrorHandler
	}
)

// WithTopicProducerOption sets the topic for the producer
func WithTopicProducerOption(topic string) ProducerOption {
	return func(p *Producer) { p.config.Topic = topic }
}

// WithBalancerProducerOption sets the balancer for Kafka Producer
func WithBalancerProducerOption(bal kafgo.Balancer) ProducerOption {
	return func(p *Producer) { p.config.Balancer = bal }
}

// WithMaxAttemptsProducerOption sets the number of tries/attempts the
// kafka producer will try before giving up
func WithMaxAttemptsProducerOption(attempts int) ProducerOption {
	return func(p *Producer) { p.config.MaxAttempts = attempts }
}

// WithQueueCapacityProducerOption sets the internal buffer capacity
// used to cache incoming messages before publishing on kafka
func WithQueueCapacityProducerOption(qc int) ProducerOption {
	return func(p *Producer) { p.config.QueueCapacity = qc }
}

// WithEncoderProducerOption encodes the message passed
// onto endpoint in desired format
func WithEncoderProducerOption(fn Encoder) ProducerOption {
	return func(p *Producer) { p.enc = fn }
}

// WithBeforesProducerOption sets before functions for the
// producer, befores are triggered before the message is
// emitted on kafka
func WithBeforesProducerOption(fns ...BeforeFunc) ProducerOption {
	return func(p *Producer) {
		p.befores = append(p.befores, fns...)
	}
}

// WithAfterProducerOption sets the after functions which are executed
// after the message is published on the kafka
func WithAfterProducerOption(fns ...AfterFunc) ProducerOption {
	return func(p *Producer) {
		p.afters = append(p.afters, fns...)
	}
}

// Endpoint returns a usable endpoint
func (p *Producer) Endpoint() endpoint.Endpoint {
	return func(
		cx context.Context,
		rqi interface{},
	) (interface{}, error) {
		// encode
		msg, err := p.enc(cx, rqi)
		if err != nil {
			err = errors.Wrap(
				err, "encode msg failed",
			)
			p.errFn(cx, msg, err)
			p.errHn.Handle(cx, err)
			return nil, err
		}

		// excute before funcs
		for _, fn := range p.befores {
			cx = fn(cx, msg)
		}

		// publsih on the kafka queue
		err = p.writer.WriteMessages(cx, msg)
		if err != nil {
			err = errors.Wrap(
				err, "write on kafka failed",
			)

			p.errFn(cx, msg, err)
			p.errHn.Handle(cx, err)
			return nil, err
		}

		// aflter funcs
		for _, fn := range p.afters {
			cx = fn(cx, msg, rqi)
		}

		// return msg
		return msg, err
	}
}

// NewProducer returns a new kafka producer
func NewProducer(
	brokers []string,
	logger log.Logger,
	options ...ProducerOption,
) (*Producer, error) {
	cfg := kafgo.WriterConfig{
		Brokers:       brokers,
		Topic:         defaultTopic,
		Balancer:      &kafgo.LeastBytes{},
		MaxAttempts:   10,
		QueueCapacity: 100,
		BatchSize:     100,
		BatchBytes:    1048576,
	}

	pr := &Producer{
		config: &cfg,
		writer: nil,
	}

	// execute options
	for _, o := range options {
		o(pr)
	}

	if pr.enc == nil {
		return nil, errors.Wrap(
			ErrCreatingConsumer, "encoder is nil",
		)
	}

	if pr.errFn == nil {
		pr.errFn = defaultErrorFunc
	}

	if pr.errHn == nil {
		pr.errHn = transport.NewLogErrorHandler(logger)
	}

	pr.writer = kafgo.NewWriter(*pr.config)
	return pr, nil
}
