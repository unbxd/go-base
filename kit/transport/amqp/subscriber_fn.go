package amqp

import (
	"context"

	kita "github.com/go-kit/kit/transport/amqp"
	mqp "github.com/streadway/amqp"
)

func NoOpResponseHandler(context.Context, *mqp.Publishing, interface{}) error {
	return nil
}

func NoOpErrorEncoder(context.Context, error, *mqp.Delivery, kita.Channel, *mqp.Publishing) {
}
