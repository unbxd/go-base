package rl

import (
	"context"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	"github.com/sethvargo/go-limiter"
	"github.com/sethvargo/go-redisstore"
	"github.com/unbxd/go-base/base/endpoint"
)

type (
	Limit struct {
		connection string
		db         int
		ratePerMin int
		store      limiter.Store
		afterFun   LimiterAfterFunc
	}

	//Key is an interface to extract key to limit on
	Key interface {
		Key() string
	}

	LimiterAfterFunc func(req interface{}, res interface{}, err error)

	Tokens struct {
		Remaining uint64
	}

	LimiterOption func(*Limit)
)

// WithRedisConnection adds a redis connection
func WithRedisConnection(conn string) LimiterOption {
	return func(l *Limit) {
		l.connection = conn
	}
}

// WithRedisDB specifies a db
func WithRedisDB(db int) LimiterOption {
	return func(l *Limit) {
		l.db = db
	}
}

// WithRateLimitPerMin sets a ratelimit per minute
func WithRateLimitPerMin(cnt int) LimiterOption {
	return func(l *Limit) {
		l.ratePerMin = cnt
	}
}

// WithLimiterAfterFunc sets an after function
func WithLimiterAfterFunc(fn LimiterAfterFunc) LimiterOption {
	return func(l *Limit) {
		l.afterFun = fn
	}
}

func (l *Limit) Endpoint() endpoint.Endpoint {
	return func(
		cx context.Context,
		rqi interface{},
	) (rsi interface{}, err error) {

		req, chk := rqi.(Key)
		if !chk {
			return nil, errors.New("rqi not string ")
		}
		_, remaining, _, ok, err := l.store.Take(cx, req.Key())

		if err != nil {
			return nil, errors.Wrap(err, "failed while rate limiting")
		}

		if !ok {
			return nil, errors.New("no tokens remaining")
		}

		l.afterFun(rqi, rsi, err)

		return &Tokens{
			Remaining: remaining,
		}, nil
	}
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(connection string, opts ...LimiterOption) (*Limit, error) {

	rl := &Limit{
		connection: connection,
		db:         1,
		ratePerMin: 100,
	}

	for _, o := range opts {
		o(rl)
	}

	store, err := redisstore.New(
		&redisstore.Config{
			Tokens:   uint64(rl.ratePerMin),
			Interval: time.Minute,
			Dial: func() (redis.Conn, error) {
				return redis.Dial("tcp", rl.connection,
					redis.DialDatabase(rl.db),
				)
			},
		})

	if err != nil {
		return nil, errors.Wrap(err, "failed to create ratelimiter")
	}

	rl.store = store

	return rl, nil
}
