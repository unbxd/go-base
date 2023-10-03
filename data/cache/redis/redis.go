package redis

/**
	Please Note: We use github.com/redis/go-redis/v9 in this repository
	which doesn't support redis version older than 7.

	To use this implementation of cache, use redis 7+.
**/

import (
	"context"
	"fmt"
	"time"

	redis "github.com/redis/go-redis/v9"
	"github.com/unbxd/go-base/log"
)

var NOEXPIRE = time.Duration(0)

type (
	cache struct {
		logger log.Logger
		opt    *redis.Options

		cc *redis.Client
	}

	Option func(*cache)
)

func (c *cache) set(
	cx context.Context,
	key string,
	val []byte,
	duration time.Duration,
) error {
	var err error

	stcmd := c.cc.Set(cx, key, val, duration)
	err = stcmd.Err()

	return err
}

func (c *cache) Set(cx context.Context, key string, val []byte) {
	if err := c.set(cx, key, val, 365*24*time.Hour); err != nil {
		c.logger.Error(
			"failed to write to redis",
			log.String("key", key),
			log.Error(err),
		)
		return
	}
}

func (c *cache) exists(cx context.Context, key string) (bool, error) {
	var (
		err    error
		intcmd *redis.IntCmd
		rc     int64
	)

	intcmd = c.cc.Exists(cx, key)
	err = intcmd.Err()

	if err == nil && rc != 0 {
		return true, nil
	}

	return false, err
}

func (c *cache) Add(
	cx context.Context,
	key string,
	value []byte,
) error {
	exists, err := c.exists(cx, key)
	if err != nil {
		c.logger.Error(
			"failed to check exits from redis",
			log.String("key", key),
			log.Error(err),
		)
		return err
	}

	if exists {
		return fmt.Errorf("Item %s already exists", key)
	}

	// set now
	return c.set(cx, key, value, NOEXPIRE)
}

func (c *cache) delete(
	cx context.Context,
	key string,
) error {
	var (
		intcmd *redis.IntCmd
		err    error
		rc     int64
	)

	intcmd = c.cc.Del(cx, key)
	err = intcmd.Err()
	if err != nil {
		return err
	}

	rc, err = intcmd.Result()
	if err != nil {
		return err
	}

	if rc == 0 {
		return fmt.Errorf("no item deleted for key: %s", key)
	}

	return nil
}

func (c *cache) Replace(
	cx context.Context,
	key string,
	value []byte,
) error {
	exists, err := c.exists(cx, key)
	if err != nil {
		c.logger.Error(
			"failed to check exits from redis",
			log.String("key", key),
			log.Error(err),
		)
	}

	if exists {
		err = c.delete(cx, key)
		if err != nil {
			c.logger.Error(
				"failed to delete key from redis",
				log.String("key", key),
				log.Error(err),
			)
		}
	}

	// set the data
	return c.set(cx, key, value, NOEXPIRE)
}

func (c *cache) SetWithDuration(
	cx context.Context,
	key string,
	val []byte,
	expiration time.Duration,
) {
	err := c.set(cx, key, val, expiration)
	if err != nil {
		c.logger.Error(
			"failed to write to redis",
			log.String("key", key),
			log.Error(err),
		)
		return
	}
}

func (c *cache) Get(cx context.Context, key string) (val []byte, found bool) {
	var (
		strcmd *redis.StringCmd
		err    error
	)

	strcmd = c.cc.Get(cx, key)
	err = strcmd.Err()

	if err != nil && err == redis.Nil {
		return nil, false
	}

	if err != nil {
		c.logger.Error(
			"failed to get data from redis",
			log.String("key", key),
			log.Error(err),
		)

		return nil, false
	}

	vs, err := strcmd.Result()
	if err != nil && err == redis.Nil {
		return nil, false
	}
	if err != nil {
		c.logger.Error(
			"failed to get data from redis",
			log.String("key", key),
			log.Error(err),
		)
		return nil, false
	}

	return []byte(vs), true
}

func (c *cache) Delete(
	cx context.Context,
	key string,
) {
	err := c.delete(cx, key)
	if err != nil {
		c.logger.Error(
			"failed to delete data from redis",
			log.String("key", key),
			log.Error(err),
		)
	}
}

func WithPassword(password string) Option {
	return func(cc *cache) {
		cc.opt.Password = password
	}
}

func WithDatabase(db int) Option {
	return func(cc *cache) {
		cc.opt.DB = db
	}
}

func WithOnConnect(callback func(context.Context, *redis.Conn) error) Option {
	return func(cc *cache) {
		cc.opt.OnConnect = callback
	}
}

type Cache struct{ *cache }

func NewRedisCache(
	logger log.Logger,
	addr string,
	options ...Option,
) (*Cache, error) {
	opt := &redis.Options{
		Addr: addr,
	}

	ch := &cache{logger: logger, opt: opt, cc: nil}

	for _, fn := range options {
		fn(ch)
	}

	// create client
	cc := redis.NewClient(ch.opt)

	ch.cc = cc
	return &Cache{ch}, nil
}
