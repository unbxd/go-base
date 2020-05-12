package nats

import (
	"context"
	kitep "github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/transport"
	kitn "github.com/go-kit/kit/transport/nats"
	natn "github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"github.com/unbxd/go-base/base/endpoint"
	"github.com/unbxd/go-base/base/log"
)

type (

	// Decoder decodes the message received on NATS and converts into business entity
	Decoder func(context.Context, *natn.Msg) (request interface{}, err error)

	// ResponseHandler handles the endpoint response
	ResponseHandler func(context.Context, string, *natn.Conn, interface{}) error

	BeforeFunc func(context.Context, *natn.Msg) context.Context

	AfterFunc func(context.Context, *natn.Conn) context.Context

	ErrorEncoder kitn.ErrorEncoder

	ErrorHandler interface{ transport.ErrorHandler }

	// Subscriber is NATS subscription
	subscriber struct {
		*kitn.Subscriber

		id       string
		subject  string
		qGroup   string
		end      endpoint.Endpoint
		dec      Decoder
		reshn    ResponseHandler
		befores  []BeforeFunc
		afters   []AfterFunc
		errorEnc ErrorEncoder
		errorhn  ErrorHandler
		conn     *natn.Conn

		middlewares []endpoint.Middleware

		subscription *natn.Subscription
		options      []kitn.SubscriberOption
	}

	// SubscriberOption provides set of options to modify a Subscriber
	SubscriberOption func(*subscriber)
)

func (s *subscriber) Id() string {
	return s.id
}

func (s *subscriber) Topic() string {
	return s.subject
}

func (s *subscriber) Group() string {
	return s.qGroup
}

func (s *subscriber) IsValid() bool {
	return s.subscription != nil && s.subscription.IsValid()
}

func WithQGroupSubscriberOption(qGroup string) SubscriberOption {
	return func(s *subscriber) {
		s.qGroup = qGroup
	}
}

func WithId(id string) SubscriberOption {
	return func(s *subscriber) {
		s.id = id
	}
}

func WithSubjectSubscriberOption(sub string) SubscriberOption {
	return func(s *subscriber) {
		s.subject = sub
	}
}

func WithEndpointSubscriberOption(end endpoint.Endpoint) SubscriberOption {
	return func(s *subscriber) {
		s.end = end
	}
}

func WithDecoderSubscriberOption(fn Decoder) SubscriberOption {
	return func(s *subscriber) {
		s.dec = fn
	}
}

func WithBeforeFuncsSubscriberOption(fns ...BeforeFunc) SubscriberOption {
	return func(s *subscriber) {
		s.befores = append(s.befores, fns...)
		for _, fn := range fns {
			s.options = append(
				s.options,
				kitn.SubscriberBefore(kitn.RequestFunc(fn)),
			)
		}
	}
}

func WithAfterFuncsSubscriberOption(fns ...AfterFunc) SubscriberOption {
	return func(s *subscriber) {
		s.afters = append(s.afters, fns...)
		for _, fn := range fns {
			s.options = append(
				s.options,
				kitn.SubscriberAfter(kitn.SubscriberResponseFunc(fn)),
			)
		}
	}
}

// HandlerWithEndpointMiddleware provides an ability to add a
// middleware of the base type
func WithEndpointMiddleware(m endpoint.Middleware) SubscriberOption {
	return func(s *subscriber) {
		s.middlewares = append(s.middlewares, m)
	}
}

func WithErrorEncoderSubscriberOption(e ErrorEncoder) SubscriberOption {
	return func(s *subscriber) {
		s.errorEnc = e
		s.options = append(
			s.options,
			kitn.SubscriberErrorEncoder(kitn.ErrorEncoder(e)),
		)
	}
}

func WithErrorhandlerSubscriberOption(e ErrorHandler) SubscriberOption {
	return func(s *subscriber) {
		s.errorhn = e
		s.options = append(s.options, kitn.SubscriberErrorHandler(e))
	}
}

func (s *subscriber) open() error {

	var err error
	if len(s.qGroup) > 0 {
		s.subscription, err = s.conn.QueueSubscribe(
			s.subject,
			s.qGroup,
			s.ServeMsg(s.conn),
		)
	} else {
		s.subscription, err = s.conn.Subscribe(
			s.subject,
			s.ServeMsg(s.conn),
		)
	}

	return err
}

func (s *subscriber) close() error {
	return s.subscription.Drain()
}

func newSubscriber(
	logger log.Logger,
	con *natn.Conn,
	options ...SubscriberOption,
) (*subscriber, error) {

	s := subscriber{conn: con}

	for _, o := range options {
		o(&s)
	}

	if s.end == nil {
		return nil, errors.Wrap(
			ErrCreatingSubscriber, "missing endpoint",
		)
	}

	if len(s.subject) == 0 {
		return nil, errors.Wrap(
			ErrCreatingSubscriber, "missing subject",
		)
	}

	if s.dec == nil {
		return nil, errors.Wrap(
			ErrCreatingSubscriber, "missing decoder",
		)
	}

	if s.reshn == nil {
		s.reshn = NoOpResponseHandler
	}

	if s.errorEnc == nil {
		WithErrorEncoderSubscriberOption(NoOpErrorEncoder)
	}

	if s.errorhn == nil {
		WithErrorhandlerSubscriberOption(transport.NewLogErrorHandler(logger))
	}

	s.Subscriber = kitn.NewSubscriber(
		kitep.Endpoint(
			wrap(s.end, s.middlewares...),
		),
		kitn.DecodeRequestFunc(s.dec),
		kitn.EncodeResponseFunc(s.reshn),
		s.options...,
	)

	return &s, nil
}

func wrap(ep endpoint.Endpoint, mws ...endpoint.Middleware) endpoint.Endpoint {

	newmw := endpoint.Chain(
		noopMiddleware,
		mws...,
	)

	return newmw(ep)
}

func noopMiddleware(next endpoint.Endpoint) endpoint.Endpoint {
	return func(
		ctx context.Context,
		req interface{},
	) (interface{}, error) {
		return next(ctx, req)
	}
}
