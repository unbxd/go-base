package grpc

import (
	"fmt"
	"net"

	grpc "google.golang.org/grpc"
)

type (
	// TransportOption sets options parameters to Transport
	TransportOption func(*Transport)

	// Transport is a wrapper around grpc.Server
	Transport struct {
		*grpc.Server
		Port int
	}
)

// Open starts the Transport
func (tr *Transport) Open() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", tr.Port))
	if err != nil {
		return err
	}
	return tr.Serve(listener)
}

// NewTransport returns a new transport
func NewTransport(options ...TransportOption) (*Transport, error) {
	transport := &Transport{
		Server: grpc.NewServer(),
	}
	for _, o := range options {
		o(transport)
	}
	return transport, nil
}
