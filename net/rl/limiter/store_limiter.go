package limiter

import (
	"context"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	"github.com/sethvargo/go-limiter"
	"github.com/sethvargo/go-redisstore"
)

type (
	store struct {
		connection  string
		db          int
		ratePerMin  int
		store       limiter.Store
		errCallback func(error) bool
	}
	Option func(*store)
)

func (l *store) Allow(cx context.Context, key string) (bool, error) {
	_, _, _, ok, err := l.store.Take(cx, key)
	if err != nil {
		return l.errCallback(err), nil
	}

	return ok, nil
}

// WithRedisDB sets a default db
func WithRedisDB(db int) Option {
	return func(s *store) {
		s.db = db
	}
}

// WithRateLimitPerMin sets token limit per min
func WithRateLimitPerMin(cnt int) Option {
	return func(s *store) {
		s.ratePerMin = cnt
	}
}

func WithErrCallback(fn func(error) bool) Option {
	return func(s *store) {
		s.errCallback = fn
	}
}

// NewStoreLimiter creates a redis backed limiter
func NewStoreLimiter(conn string, opts ...Option) (Limiter, error) {
	store := &store{
		connection: conn,
		db:         1,
		ratePerMin: 100,
		errCallback: func(err error) bool {
			return true
		},
	}

	for _, o := range opts {
		o(store)
	}

	red_st, err := redisstore.New(
		&redisstore.Config{
			Tokens:   uint64(store.ratePerMin),
			Interval: time.Minute,
			Dial: func() (redis.Conn, error) {
				return redis.Dial("tcp", store.connection,
					redis.DialDatabase(store.db),
				)
			},
		})

	if err != nil {
		return nil, errors.Wrap(err, "failed to create redis limiter")
	}

	store.store = red_st

	return store, nil
}
