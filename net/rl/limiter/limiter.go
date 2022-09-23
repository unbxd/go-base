package limiter

import "context"

type Limiter interface {
	Allow(context.Context, string) (bool, error)
}
