package nats

import (
	"context"
	"time"

	natn "github.com/nats-io/nats.go"
	"github.com/pkg/errors"
)

type (
	ConsumerOption func(*Consumer)

	// ConsumerErrorHandler is a function that is called when an error occurs
	ConsumerErrorHandler func(context.Context, error) error

	Consumer struct {
		conn *natn.Conn
		opts *natn.Options
		cfg  *natn.ConsumerConfig
		info *natn.ConsumerInfo
		js   natn.JetStreamContext

		errorHandler ConsumerErrorHandler
	}
)

func WithDurable(durable string) ConsumerOption {
	return func(c *Consumer) {
		c.cfg.Durable = durable
	}
}

func WithDeliverSubject(deliverSubject string) ConsumerOption {
	return func(c *Consumer) {
		c.cfg.DeliverSubject = deliverSubject
	}
}

func WithDeliverGroup(deliverGroup string) ConsumerOption {
	return func(c *Consumer) {
		c.cfg.DeliverGroup = deliverGroup
	}
}

func WithDeliverPolicy(deliverPolicy natn.DeliverPolicy) ConsumerOption {
	return func(c *Consumer) {
		c.cfg.DeliverPolicy = deliverPolicy
	}
}

func WithOptStartSeq(optStartSeq uint64) ConsumerOption {
	return func(c *Consumer) {
		c.cfg.OptStartSeq = optStartSeq
	}
}

func WithOptStartTime(optStartTime *time.Time) ConsumerOption {
	return func(c *Consumer) {
		c.cfg.OptStartTime = optStartTime
	}
}

func WithAckPolicy(ackPolicy natn.AckPolicy) ConsumerOption {
	return func(c *Consumer) {
		c.cfg.AckPolicy = ackPolicy
	}
}

func WithAckWait(ackWait time.Duration) ConsumerOption {
	return func(c *Consumer) {
		c.cfg.AckWait = ackWait
	}
}

func WithMaxDeliver(maxDeliver int) ConsumerOption {
	return func(c *Consumer) {
		c.cfg.MaxDeliver = maxDeliver
	}
}

func WithFilterSubject(filterSubject string) ConsumerOption {
	return func(c *Consumer) {
		c.cfg.FilterSubject = filterSubject
	}
}

func WithReplayPolicy(replayPolicy natn.ReplayPolicy) ConsumerOption {
	return func(c *Consumer) {
		c.cfg.ReplayPolicy = replayPolicy
	}
}

func WithRateLimit(rateLimit uint64) ConsumerOption {
	return func(c *Consumer) {
		c.cfg.RateLimit = rateLimit
	}
}

func WithSampleFrequency(sampleFrequency string) ConsumerOption {
	return func(c *Consumer) {
		c.cfg.SampleFrequency = sampleFrequency
	}
}

func WithMaxAckPending(maxAckPending int) ConsumerOption {
	return func(c *Consumer) {
		c.cfg.MaxAckPending = maxAckPending
	}
}

func WithFlowControl(flowControl bool) ConsumerOption {
	return func(c *Consumer) {
		c.cfg.FlowControl = flowControl
	}
}

func WithHeartbeat(heartbeat time.Duration) ConsumerOption {
	return func(c *Consumer) {
		c.cfg.Heartbeat = heartbeat
	}
}

func WithMaxRequestBatch(maxRequestBatch int) ConsumerOption {
	return func(c *Consumer) {
		c.cfg.MaxRequestBatch = maxRequestBatch
	}
}

func WithMaxRequestExpires(maxRequestExpires time.Duration) ConsumerOption {
	return func(c *Consumer) {
		c.cfg.MaxRequestExpires = maxRequestExpires
	}
}

func WithInactiveThreshold(inactiveThreshold time.Duration) ConsumerOption {
	return func(c *Consumer) {
		c.cfg.InactiveThreshold = inactiveThreshold
	}
}

func defaultConsumerErrorHandler(cx context.Context, err error) error {
	return err
}

func NewConsumer(connstr string, stream string, options ...ConsumerOption) (*Consumer, error) {
	var (
		err  error
		cc   *natn.Conn
		opts = natn.GetDefaultOptions()
		c    = &Consumer{
			conn:         nil,
			opts:         &opts,
			cfg:          &natn.ConsumerConfig{},
			info:         &natn.ConsumerInfo{},
			errorHandler: defaultConsumerErrorHandler,
		}
	)

	for _, fn := range options {
		fn(c)
	}

	c.opts.Url = connstr

	// Connect to NATS
	cc, err = c.opts.Connect()

	if err != nil {
		return nil, errors.Wrap(
			err, "unable to connect to nats server",
		)
	}

	// Create JetStream Context
	js, err := cc.JetStream()

	if err != nil {
		return nil, errors.Wrap(
			err, "unable to create JetStream context",
		)
	}

	// Create a consumer
	info, err := js.AddConsumer(stream, c.cfg)

	if err != nil {
		return nil, errors.Wrap(err, "unable to create consumer")
	}

	c.js = js
	c.conn = cc
	c.info = info

	return c, nil
}

func (c *Consumer) UpdateConsumer(stream string, options ...ConsumerOption) (*Consumer, error) {
	for _, fn := range options {
		fn(c)
	}

	// Update the consumer
	info, err := c.js.UpdateConsumer(stream, c.cfg)

	if err != nil {
		return nil, errors.Wrap(err, "unable to update consumer")
	}

	c.info = info

	return c, nil
}

func (c *Consumer) DeleteConsumer(stream string) error {

	// Delete the consumer
	err := c.js.DeleteConsumer(stream, c.cfg.Durable)

	if err != nil {
		return errors.Wrap(err, "unable to delete consumer")
	}

	return nil
}
