package nats

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/nats-io/nats.go"
	natn "github.com/nats-io/nats.go"
	"github.com/unbxd/go-base/v2/errors"
)

// EncodeJSONRequest is an Encoder that serializes the request as a
// JSON object to the Data of the Msg. Many JSON-over-NATS services can use it as
// a sensible default.
func EncodeJSONRequest(_ context.Context, subject string, request interface{}) (*nats.Msg, error) {
	var (
		buf bytes.Buffer
		err error
	)

	err = json.NewEncoder(&buf).Encode(request)
	if err != nil {
		return nil, errors.Wrap(err, "defaultencoder: encoding error")
	}

	return &natn.Msg{
		Subject: subject,
		Data:    buf.Bytes(),
	}, err
}
