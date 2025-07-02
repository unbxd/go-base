package rate

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	maxRetries = 3
	minTTL     = 60 * time.Second
	ttlBuffer  = 10 * time.Second
)

// redisLimiter implements a Redis-backed distributed rate limiter using the token bucket algorithm.
type redisLimiter struct {
	client redis.UniversalClient
	limit  float64 // tokens per second
	burst  int     // maximum burst size
}

// NewRedisLimiter creates a new Redis-backed rate limiter.
func NewRedisLimiter(client redis.UniversalClient, limit float64, burst int) Limiter {
	return &redisLimiter{
		client: client,
		limit:  limit,
		burst:  burst,
	}
}

// Allow reports whether an event may happen now for the given key.
func (r *redisLimiter) Allow(key Key) bool {
	if r.limit <= 0 {
		return false
	}

	ctx := context.Background()
	redisKey := r.formatKey(key)
	now := time.Now().UnixNano()
	ttl := r.calculateTTL()

	for i := 0; i < maxRetries; i++ {
		allowed, err := r.tryAllow(ctx, redisKey, now, ttl)
		if err == nil {
			return allowed
		}
		if err == redis.TxFailedErr {
			continue // Retry on transaction conflict
		}
		return false // Fail closed on other errors
	}

	return false // Fail closed after max retries
}

// formatKey creates a Redis key for the given rate limiter key.
func (r *redisLimiter) formatKey(key Key) string {
	return fmt.Sprintf("rate:limiter:%s", key)
}

// tryAllow attempts to consume 1 token atomically using Redis transactions.
func (r *redisLimiter) tryAllow(ctx context.Context, redisKey string, now int64, ttl time.Duration) (bool, error) {
	var allowed bool

	err := r.client.Watch(ctx, func(tx *redis.Tx) error {
		tokens, last, err := r.getCurrentState(ctx, tx, redisKey, now)
		if err != nil {
			return err
		}

		newTokens := r.calculateTokens(tokens, last, now)
		allowed = newTokens >= 1.0
		if allowed {
			newTokens--
		}

		return r.updateState(ctx, tx, redisKey, newTokens, now, ttl)
	}, redisKey)

	return allowed, err
}

// getCurrentState retrieves the current token count and last update time from Redis.
func (r *redisLimiter) getCurrentState(ctx context.Context, tx *redis.Tx, redisKey string, now int64) (float64, int64, error) {
	vals, err := tx.HMGet(ctx, redisKey, "tokens", "last").Result()
	if err != nil && err != redis.Nil {
		return 0, 0, err
	}

	tokens := float64(r.burst) // Default to full burst
	last := now                // Default to current time

	if len(vals) >= 2 {
		if vals[0] != nil {
			if tokensStr, ok := vals[0].(string); ok {
				if parsed, err := strconv.ParseFloat(tokensStr, 64); err == nil {
					tokens = parsed
				}
			}
		}
		if vals[1] != nil {
			if lastStr, ok := vals[1].(string); ok {
				if parsed, err := strconv.ParseInt(lastStr, 10, 64); err == nil {
					last = parsed
				}
			}
		}
	}

	return tokens, last, nil
}

// calculateTokens computes the current token count after refill.
func (r *redisLimiter) calculateTokens(currentTokens float64, lastUpdate, now int64) float64 {
	deltaSeconds := float64(now-lastUpdate) / float64(time.Second)
	if deltaSeconds < 0 {
		deltaSeconds = 0
	}

	refill := deltaSeconds * r.limit
	return math.Min(float64(r.burst), currentTokens+refill)
}

// updateState atomically updates the token count and timestamp in Redis.
func (r *redisLimiter) updateState(ctx context.Context, tx *redis.Tx, redisKey string, tokens float64, now int64, ttl time.Duration) error {
	pipe := tx.TxPipeline()
	pipe.HMSet(ctx, redisKey, map[string]interface{}{
		"tokens": fmt.Sprintf("%.9f", tokens),
		"last":   strconv.FormatInt(now, 10),
	})
	pipe.Expire(ctx, redisKey, ttl)

	_, err := pipe.Exec(ctx)
	return err
}

// calculateTTL calculates an appropriate TTL for the Redis key.
func (r *redisLimiter) calculateTTL() time.Duration {
	if r.limit <= 0 {
		return minTTL
	}

	// TTL should be at least the time to fill the bucket plus buffer
	fillTime := time.Duration(float64(r.burst)/r.limit) * time.Second
	ttl := fillTime + ttlBuffer

	if ttl < minTTL {
		ttl = minTTL
	}

	return ttl
}
