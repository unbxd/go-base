package kafka

import (
	"context"
	"fmt"

	kafgo "github.com/segmentio/kafka-go"
)

func defaultErrorFunc(
	ctx context.Context,
	msg kafgo.Message,
	err error,
) {
	//nolint:forbidigo
	fmt.Printf(
		"Reader Err: [Topic: %v] [Partition: %v] [Key: %v] [Error: %v]",
		msg.Topic,
		msg.Partition,
		msg.Key,
		err.Error(),
	)
}
