package nats

import (
	"context"
	"encoding/json"
	natn "github.com/nats-io/nats.go"
)

// EncodeJSONRequest is an Encoder that serializes the request as a
// JSON object to the Data of the Msg. Many JSON-over-NATS services can use it as
// a sensible default.
func EncodeJSONRequest(_ context.Context, msg *natn.Msg, request interface{}) error {
	b, err := json.Marshal(request)
	if err != nil {
		return err
	}

	msg.Data = b

	return nil
}
