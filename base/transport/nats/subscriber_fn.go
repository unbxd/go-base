package nats

import (
	"context"
	natn "github.com/nats-io/nats.go"
)

func NoOpResponseHandler(context.Context, string, *natn.Conn, interface{}) error {
	return nil
}

func NoOpErrorEncoder(context.Context, error, string, *natn.Conn) {
	return
}
