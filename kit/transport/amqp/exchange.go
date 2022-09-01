package amqp

import (
	"github.com/pkg/errors"
	mqp "github.com/streadway/amqp"
)

type (
	exchange struct {
		name       string
		kind       string
		durable    bool
		autoDelete bool
		internal   bool
		noWait     bool
		args       mqp.Table
	}

	// ExchangeOption provides set of options to modify a Exchange
	ExchangeOption func(*exchange)
)

func NewExchange(name string, opts ...ExchangeOption) (*exchange, error) {

	// not using the default exchange
	if name == "" {
		return nil, errors.Wrap(ErrCreatingExchange, "exchange name is empty")
	}

	// setting default values for exchange
	e := &exchange{
		name:       name,
		kind:       "direct",
		durable:    false,
		autoDelete: false,
		internal:   false,
		noWait:     false,
		args:       nil,
	}

	// default settings can be overwritten by options
	for _, option := range opts {
		option(e)
	}

	return e, nil
}

func WithKindExchangeOption(kind string) ExchangeOption {
	return func(e *exchange) {
		e.kind = kind
	}
}

func WithDurableExchangeOption(durable bool) ExchangeOption {
	return func(e *exchange) {
		e.durable = durable
	}
}

func WithAutoDeleteExchangeOption(autoDelete bool) ExchangeOption {
	return func(e *exchange) {
		e.autoDelete = autoDelete
	}
}

func WithInternalExchangeOption(internal bool) ExchangeOption {
	return func(e *exchange) {
		e.internal = internal
	}
}

func WithNoWaitExchangeOption(noWait bool) ExchangeOption {
	return func(e *exchange) {
		e.noWait = noWait
	}
}

func WithArgsExchangeOption(args mqp.Table) ExchangeOption {
	return func(e *exchange) {
		e.args = args
	}
}
