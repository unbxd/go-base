package amqp

import (
	"context"

	kitep "github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/transport"
	kita "github.com/go-kit/kit/transport/amqp"
	"github.com/pkg/errors"
	mqp "github.com/streadway/amqp"
	"github.com/unbxd/go-base/kit/endpoint"
	"github.com/unbxd/go-base/utils/log"
)

type (
	// Decoder decodes the message received on RabbitMQ and converts into business entity
	Decoder func(context.Context, *mqp.Delivery) (request interface{}, err error)

	// ResponseHandler handles the endpoint response
	ResponseHandler func(context.Context, *mqp.Publishing, interface{}) error

	BeforeFunc func(context.Context, *mqp.Publishing, *mqp.Delivery) context.Context

	AfterFunc func(context.Context, *mqp.Delivery, kita.Channel, *mqp.Publishing) context.Context

	ErrorEncoder func(ctx context.Context, err error, deliv *mqp.Delivery, ch kita.Channel, pub *mqp.Publishing)

	ErrorHandler interface{ transport.ErrorHandler }

	subscriber struct {
		*kita.Subscriber

		conn *mqp.Connection

		id       string
		channel  *mqp.Channel
		queue    *mqp.Queue
		q        string
		exchange string

		end      endpoint.Endpoint
		dec      Decoder
		reshn    ResponseHandler
		befores  []BeforeFunc
		afters   []AfterFunc
		errorEnc ErrorEncoder
		errorhn  ErrorHandler

		middlewares []endpoint.Middleware

		options []kita.SubscriberOption
	}

	// SubscriberOption provides set of options to modify a Subscriber
	SubscriberOption func(*subscriber)
)

func (s *subscriber) Id() string {
	return s.id
}

func (s *subscriber) Queue() string {
	return s.q
}

func (s *subscriber) Exchange() string {
	return s.exchange
}

func WithId(id string) SubscriberOption {
	return func(s *subscriber) {
		s.id = id
	}
}

func WithQSubscriberOption(q string) SubscriberOption {
	return func(s *subscriber) {
		s.q = q
	}
}

func WithExchangeSubscriberOption(exchange string) SubscriberOption {
	return func(s *subscriber) {
		s.exchange = exchange
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
				kita.SubscriberBefore(kita.RequestFunc(fn)),
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
				kita.SubscriberAfter(kita.SubscriberResponseFunc(fn)),
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
			kita.SubscriberErrorEncoder(kita.ErrorEncoder(e)),
		)
	}
}

func WithErrorhandlerSubscriberOption(e ErrorHandler) SubscriberOption {
	return func(s *subscriber) {
		s.errorhn = e
		s.options = append(s.options, kita.SubscriberErrorHandler(e))
	}
}

//TODO: Check if this works
func (s *subscriber) open() error {
	var err error

	s.ServeDelivery(s.channel)(&mqp.Delivery{})

	return err
}

func (s *subscriber) close() error {
	return s.channel.Close()
}

func newSubscriber(
	logger log.Logger,
	conn *mqp.Connection,
	options ...SubscriberOption,
) (*subscriber, error) {

	s := subscriber{conn: conn}

	for _, o := range options {
		o(&s)
	}

	if s.end == nil {
		return nil, errors.Wrap(
			ErrCreatingSubscriber, "missing endpoint",
		)
	}

	if s.exchange == "" {
		return nil, errors.Wrap(
			ErrCreatingSubscriber, "missing exchange",
		)
	}

	if s.q == "" {
		return nil, errors.Wrap(
			ErrCreatingSubscriber, "missing queue",
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

	ch, err := s.conn.Channel()
	if err != nil {
		return nil, errors.Wrap(ErrCreatingSubscriber, "creating channel failed")
	}
	s.channel = ch

	err = s.channel.ExchangeDeclare(
		s.exchange,
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, errors.Wrap(ErrCreatingSubscriber, "declaring exchange failed")
	}

	queue, err := s.channel.QueueDeclare(
		s.q,
		true,
		false,
		true,
		false,
		nil,
	)
	if err != nil {
		return nil, errors.Wrap(ErrCreatingSubscriber, "declaring queue failed")
	}
	s.queue = &queue

	err = s.channel.QueueBind(
		queue.Name,
		"",
		s.exchange,
		false,
		nil,
	)
	if err != nil {
		return nil, errors.Wrap(ErrCreatingSubscriber, "failed to bind a queue")
	}

	s.Subscriber = kita.NewSubscriber(
		kitep.Endpoint(
			wrap(s.end, s.middlewares...),
		),
		kita.DecodeRequestFunc(s.dec),
		kita.EncodeResponseFunc(s.reshn),
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
