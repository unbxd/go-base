package nats

import (
	natn "github.com/nats-io/nats.go"
	"github.com/unbxd/go-base/base/log"
)

// disconnectErrorCallback is called when the connection to
// nats server is lost
func disconnectErrorCallback(logger log.Logger) natn.ConnErrHandler {
	return func(nc *natn.Conn, err error) {
		logger.Error("disconnected from nats", log.Error(err))
	}
}

// reconnectCallback is called when the connection to nats
// server is re-established
func reconnectCallback(logger log.Logger) natn.ConnHandler {
	return func(nc *natn.Conn) {
		logger.Info("Got reconnected", log.String("url", nc.ConnectedUrl()))
	}
}
