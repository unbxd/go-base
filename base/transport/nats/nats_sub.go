package nats

import (
	"bytes"
	"context"
	"sync"

	"github.com/pkg/errors"

	"encoding/json"

	"github.com/go-kit/kit/endpoint"
	kitn "github.com/go-kit/kit/transport/nats"
	natn "github.com/nats-io/nats.go"
)

// ContextKey is key for Context
type ContextKey int

// Keys for message
const (
	ContextKeySubject ContextKey = iota
)

// Request defines the standard JSON request. For more specialized usecase
// use NewCustomSubscriber()
type Request map[string]interface{}

// Response defines the standard JSON response. For more specialized
// use case, use NewCustomSubscriber()
type Response map[string]interface{}

// Handler exposes a method through which we can read a message from
// nats server and respond with a Response of our own
type Handler func(context.Context, Request) (Response, error)

// Subscriber is a wrapper on top of go-kit/transport/nats.Subscriber
type Subscriber struct {
	*kitn.Subscriber

	subject      string
	waitgroup    *sync.WaitGroup
	subscription *natn.Subscription
	callback     natn.MsgHandler
}

// Subscribe subscribes the subscriber to the subject
func (s *Subscriber) Subscribe(connection *natn.Conn) (err error) {
	defer s.waitgroup.Add(1)

	if s.callback == nil {
		// subscriber is not of simple type
		s.callback = s.ServeMsg(connection)
	}

	sub, err := connection.Subscribe(s.subject, s.callback)
	if err != nil {
		return
	}

	s.subscription = sub
	return
}

// Close closes the subscriber
func (s *Subscriber) Close() (err error) {
	s.waitgroup.Done()
	return
}

// DecodeJSONRequest decodes natn.Msg and converts into Request
func DecodeJSONRequest(ctx context.Context, msg *natn.Msg) (req interface{}, err error) {
	var request Request

	err = json.NewDecoder(bytes.NewReader(msg.Data)).Decode(request)
	if err != nil {
		return
	}

	return request, err
}

func NewSimpleSubscriber(
	subject string,
	waitgroup *sync.WaitGroup,
	callback natn.MsgHandler,
) (sub *Subscriber, err error) {
	sub = &Subscriber{
		subject:   subject,
		waitgroup: waitgroup,
		callback:  callback,
	}

	return
}

func wrap(hn Handler) endpoint.Endpoint {
	return func(ctx context.Context, req interface{}) (res interface{}, err error) {
		r, ok := req.(Request)
		if !ok {
			return nil, errors.New("request is not Request")
		}
		return hn(ctx, r)
	}
}

func NewSubscriber(
	subject string,
	waitgroup *sync.WaitGroup,
	handler Handler,
	options ...kitn.SubscriberOption,
) (sub *Subscriber, err error) {
	sub = &Subscriber{
		Subscriber: kitn.NewSubscriber(
			wrap(handler),
			DecodeJSONRequest,
			kitn.EncodeJSONResponse,
			options...,
		),
		subject:   subject,
		waitgroup: waitgroup,
	}

	return
}

func NewCustomSubscriber(
	subject string,
	waitgroup *sync.WaitGroup,
	ep endpoint.Endpoint,
	dec kitn.DecodeRequestFunc,
	enc kitn.EncodeResponseFunc,
	options ...kitn.SubscriberOption,
) (sub *Subscriber, err error) {
	sub = &Subscriber{
		Subscriber: kitn.NewSubscriber(
			ep, dec, enc, options...,
		),
		subject:   subject,
		waitgroup: waitgroup,
	}

	return
}

// Subscribers holds the list of subscribers
type Subscribers struct {
	subs           []Subscriber
	defaultOptions []kitn.SubscriberOption
	waitgroup      *sync.WaitGroup
}
