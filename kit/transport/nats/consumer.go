package nats

import (
	"time"

	natn "github.com/nats-io/nats.go"
)

type (
	ConsumerOption func(*natn.ConsumerConfig)
)

func WithDurable(durable string) ConsumerOption {
	return func(c *natn.ConsumerConfig) {
		c.Durable = durable
	}
}

func WithDeliverSubject(deliverSubject string) ConsumerOption {
	return func(c *natn.ConsumerConfig) {
		c.DeliverSubject = deliverSubject
	}
}

func WithDeliverGroup(deliverGroup string) ConsumerOption {
	return func(c *natn.ConsumerConfig) {
		c.DeliverGroup = deliverGroup
	}
}

func WithDeliverPolicy(deliverPolicy natn.DeliverPolicy) ConsumerOption {
	return func(c *natn.ConsumerConfig) {
		c.DeliverPolicy = deliverPolicy
	}
}

func WithOptStartSeq(optStartSeq uint64) ConsumerOption {
	return func(c *natn.ConsumerConfig) {
		c.OptStartSeq = optStartSeq
	}
}

func WithOptStartTime(optStartTime *time.Time) ConsumerOption {
	return func(c *natn.ConsumerConfig) {
		c.OptStartTime = optStartTime
	}
}

func WithAckPolicy(ackPolicy natn.AckPolicy) ConsumerOption {
	return func(c *natn.ConsumerConfig) {
		c.AckPolicy = ackPolicy
	}
}

func WithAckWait(ackWait time.Duration) ConsumerOption {
	return func(c *natn.ConsumerConfig) {
		c.AckWait = ackWait
	}
}

func WithMaxDeliver(maxDeliver int) ConsumerOption {
	return func(c *natn.ConsumerConfig) {
		c.MaxDeliver = maxDeliver
	}
}

func WithFilterSubject(filterSubject string) ConsumerOption {
	return func(c *natn.ConsumerConfig) {
		c.FilterSubject = filterSubject
	}
}

func WithReplayPolicy(replayPolicy natn.ReplayPolicy) ConsumerOption {
	return func(c *natn.ConsumerConfig) {
		c.ReplayPolicy = replayPolicy
	}
}

func WithRateLimit(rateLimit uint64) ConsumerOption {
	return func(c *natn.ConsumerConfig) {
		c.RateLimit = rateLimit
	}
}

func WithSampleFrequency(sampleFrequency string) ConsumerOption {
	return func(c *natn.ConsumerConfig) {
		c.SampleFrequency = sampleFrequency
	}
}

func WithMaxAckPending(maxAckPending int) ConsumerOption {
	return func(c *natn.ConsumerConfig) {
		c.MaxAckPending = maxAckPending
	}
}

func WithFlowControl(flowControl bool) ConsumerOption {
	return func(c *natn.ConsumerConfig) {
		c.FlowControl = flowControl
	}
}

func WithHeartbeat(heartbeat time.Duration) ConsumerOption {
	return func(c *natn.ConsumerConfig) {
		c.Heartbeat = heartbeat
	}
}

func WithMaxRequestBatch(maxRequestBatch int) ConsumerOption {
	return func(c *natn.ConsumerConfig) {
		c.MaxRequestBatch = maxRequestBatch
	}
}

func WithMaxRequestExpires(maxRequestExpires time.Duration) ConsumerOption {
	return func(c *natn.ConsumerConfig) {
		c.MaxRequestExpires = maxRequestExpires
	}
}

func WithInactiveThreshold(inactiveThreshold time.Duration) ConsumerOption {
	return func(c *natn.ConsumerConfig) {
		c.InactiveThreshold = inactiveThreshold
	}
}
