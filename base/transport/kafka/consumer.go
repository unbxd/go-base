package kafka

import (
	kafgo "github.com/segmentio/kafka-go"
)

type (
	// ConsumerOption provies set of options to modify a subscriber
	ConsumerOption func(*Consumer)

	// Consumer is kafka Consumer
	Consumer struct {
		reader *kafgo.Reader
		config *kafgo.ReaderConfig
	}
)

const (
	defaultConsumerGroupID = "go-base-consumer"
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

// NewKafkaConsumer returns kafka consumer for the given brokers
func NewKafkaConsumer(
	brokers []string,
	options ...ConsumerOption,
) (*Consumer, error) {

}
