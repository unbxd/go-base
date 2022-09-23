package rl

import (
	"context"

	"github.com/pkg/errors"
	"github.com/unbxd/go-base/kit/endpoint"
	"github.com/unbxd/go-base/net/rl/limiter"
)

type (
	Limit struct {
		enable   bool
		lim      limiter.Limiter
		fn       endpoint.Endpoint
		afterFun LimiterAfterFunc
	}

	//Key is an interface to extract key to limit on
	Key interface {
		Key() string
	}

	LimiterAfterFunc func(req interface{}, res interface{}, err error)

	LimiterOption func(*Limit)
)

// WithLimiterAfterFunc sets an after function
func WithLimiterAfterFunc(fn LimiterAfterFunc) LimiterOption {
	return func(l *Limit) {
		l.afterFun = fn
	}
}

// WithRateLimiterEnabled enables rate limiter
func WithRateLimiterEnabled(flag bool) LimiterOption {
	return func(l *Limit) {
		l.enable = flag
	}
}

func WithRateLimiter(lim limiter.Limiter) LimiterOption {
	return func(l *Limit) {
		l.lim = lim
	}
}

func (l *Limit) Endpoint() endpoint.Endpoint {
	return func(
		cx context.Context,
		rqi interface{},
	) (rsi interface{}, err error) {

		if !l.enable {
			return l.fn(cx, rqi)
		}

		key, chk := rqi.(Key)
		if !chk {
			return l.fn(cx, rqi)
		}

		ok, err := l.lim.Allow(cx, key.Key())

		if err != nil {
			return rsi, err
		}

		if !ok {
			err = errors.New("rate limit exceeded")
		}

		l.afterFun(rqi, rsi, err)

		return l.fn(cx, rqi)
	}
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(
	fn endpoint.Endpoint,
	connection string,
	opts ...LimiterOption) (*Limit, error) {
	rl := &Limit{
		enable: false,
		fn:     fn,
		lim:    limiter.NewNoopLimiter(),
	}

	for _, o := range opts {
		o(rl)
	}

	return rl, nil
}
