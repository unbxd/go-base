package amqp

import (
	"context"
	"encoding/json"

	mqp "github.com/streadway/amqp"
)

// EncodeJSONRequest is an Encoder that serializes the request as a
// JSON object to the Data of the Msg. Many JSON-over-RabbitMQ services can use it as
// a sensible default.
func EncodeJSONRequest(_ context.Context, msg *mqp.Publishing, request interface{}) error {
	b, err := json.Marshal(request)
	if err != nil {
		return err
	}

	msg.Body = b

	return nil
}
