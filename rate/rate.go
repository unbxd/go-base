package rate

import (
	"context"
	"errors"

	"github.com/unbxd/go-base/v2/endpoint"
)

var ErrRateLimitExceeded = errors.New("rate limit exceeded")

type (
	Key     string
	KeyFunc func(req any) Key

	Limiter     interface{ Allow(key Key) bool }
	LimiterFunc func(key Key) bool
)

func (f LimiterFunc) Allow(key Key) bool { return f(key) }

// NewErroringLimiter returns a middleware that returns an error if the rate
// limit is exceeded. It uses the allower to check if the rate limit is exceeded
// and the keyFunc to get the key from the request
func NewErroringLimiterMiddleware(limiter Limiter, keyFunc KeyFunc) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, req any) (any, error) {
			// Don't do anything if the allower is nil
			if limiter == nil {
				return next(ctx, req)
			}

			if !limiter.Allow(keyFunc(req)) {
				return nil, ErrRateLimitExceeded
			}

			return next(ctx, req)
		}
	}
}
