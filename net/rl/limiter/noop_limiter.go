package limiter

import "context"

type (
	nope struct {
	}
)

func (l *nope) Allow(ctx context.Context, key string) (bool, error) {
	return true, nil
}

func NewNoopLimiter() Limiter {
	return &nope{}
}
