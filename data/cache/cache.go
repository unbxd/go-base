package cache

import (
	"context"
	"time"

	"github.com/unbxd/go-base/data/cache/inmem"
	"github.com/unbxd/go-base/data/cache/redis"
	"github.com/unbxd/go-base/utils/log"
)

type Cache interface {
	// Set adds item to cache replacing existing one
	Set(
		cx context.Context,
		kye string,
		val []byte,
	)

	// Add adds item to cache only if the item doesn't exist or
	// the key has expired. It won't remove an active existing value
	Add(
		cx context.Context,
		key string,
		val []byte,
	) error

	// Replace an item if it exists
	Replace(
		cx context.Context,
		key string,
		val []byte,
	) error

	// SetWithDuration sets the key with a value for a time period
	SetWithDuration(
		cx context.Context,
		key string,
		val []byte,
		expiration time.Duration,
	)

	// Get returns the value for the key from the cache and sets found flag as
	// true or it returns false if the value is not found
	Get(
		cx context.Context,
		key string,
	) (val []byte, found bool)

	// Delete deletes the key from the cache, and doesn't do anything
	// if key is not found
	Delete(
		cx context.Context,
		key string,
	)
}

func NewInMemoryCache(
	expiry time.Duration,
	eviction time.Duration,
	options ...inmem.Option,
) (Cache, error) {
	return inmem.New(expiry, eviction, options...), nil
}

func NewRedisCache(
	logger log.Logger,
	addr string,
	options ...redis.Option,
) (Cache, error) {
	return redis.NewRedisCache(
		logger,
		addr,
		options...,
	)
}
