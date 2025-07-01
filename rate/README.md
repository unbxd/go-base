# Rate Limiter Package

This package provides rate limiting functionality for Go applications using the token bucket algorithm. It offers both in-memory and Redis-backed implementations for different use cases.

## Overview

The rate limiter package implements the token bucket algorithm, which allows for smooth rate limiting with burst support. It provides a simple `Limiter` interface that can be used with different backends.

```go
type Limiter interface {
    Allow(key Key) bool
}
```

## Implementations

### 1. In-Memory Limiter

The in-memory limiter uses `golang.org/x/time/rate` internally and stores rate limiters in memory for each key.

**Features:**
- ✅ Fast (nanosecond latency)
- ✅ No external dependencies
- ✅ Perfect for single-instance applications
- ❌ Not shared across processes
- ❌ Lost on application restart

**Usage:**
```go
import "github.com/unbxd/go-base/v2/rate"

// Create an in-memory limiter: 10 requests/second, burst of 5
limiter := rate.NewInMemoryLimiter(10.0, 5)

key := rate.Key("user:123")
if limiter.Allow(key) {
    // Request allowed
    fmt.Println("Request processed")
} else {
    // Request denied
    fmt.Println("Rate limit exceeded")
}
```

### 2. Redis Limiter

The Redis limiter provides distributed rate limiting across multiple processes using Redis as the backend storage.

**Features:**
- ✅ Distributed across multiple processes
- ✅ Persistent across application restarts
- ✅ Scales horizontally
- ✅ Atomic operations with retry logic
- ❌ Requires Redis
- ❌ Network latency (~1ms per call)

**Usage:**
```go
import (
    "github.com/redis/go-redis/v9"
    "github.com/unbxd/go-base/v2/rate"
)

// Create Redis client
client := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",
    DB:   0,
})

// Create Redis limiter: 10 requests/second, burst of 5
limiter := rate.NewRedisLimiter(client, 10.0, 5)

key := rate.Key("user:123")
if limiter.Allow(key) {
    // Request allowed
    fmt.Println("Request processed")
} else {
    // Request denied
    fmt.Println("Rate limit exceeded")
}
```

## Token Bucket Algorithm

Both implementations use the token bucket algorithm:

1. **Bucket Initialization**: Each key starts with a full bucket of `burst` tokens
2. **Token Refill**: Tokens are added at a rate of `limit` tokens per second
3. **Token Consumption**: Each `Allow()` call consumes 1 token
4. **Overflow Protection**: Tokens are capped at the `burst` limit

### Example Behavior

```go
limiter := rate.NewInMemoryLimiter(2.0, 3) // 2 tokens/sec, burst 3
key := rate.Key("example")

// Initially: 3 tokens available
limiter.Allow(key) // ✅ allowed (2 tokens left)
limiter.Allow(key) // ✅ allowed (1 token left)  
limiter.Allow(key) // ✅ allowed (0 tokens left)
limiter.Allow(key) // ❌ denied (no tokens)

// After 500ms: 1 token refilled
time.Sleep(500 * time.Millisecond)
limiter.Allow(key) // ✅ allowed (0 tokens left)
```

## Middleware Integration

The package provides middleware for use with the endpoint pattern:

```go
import (
    "github.com/unbxd/go-base/v2/endpoint"
    "github.com/unbxd/go-base/v2/rate"
)

// Create rate limiter
limiter := rate.NewInMemoryLimiter(100.0, 10)

// Key extraction function
keyFunc := func(req any) rate.Key {
    if r, ok := req.(*http.Request); ok {
        return rate.Key("user:" + r.Header.Get("User-ID"))
    }
    return rate.Key("default")
}

// Create middleware
middleware := rate.NewErroringLimiterMiddleware(limiter, keyFunc)

// Apply to endpoint
endpoint := middleware(yourEndpoint)
```

## Implementation Details

### In-Memory Limiter

```go
type inMemoryLimiter struct {
    mu       sync.RWMutex
    limit    float64
    burst    int
    limiters map[Key]*rate.Limiter
}
```

- Uses `golang.org/x/time/rate.Limiter` for each key
- Thread-safe with RWMutex
- Lazy initialization of per-key limiters
- Zero limit means no requests allowed

### Redis Limiter

```go
type redisLimiter struct {
    client redis.UniversalClient
    limit  float64 // tokens per second
    burst  int     // maximum burst size
}
```

**Redis Storage:**
- Key format: `rate:limiter:{key}`
- Fields: `tokens` (float64), `last` (int64 nanoseconds)
- Automatic expiration based on refill time

**Atomic Operations:**
- Uses `WATCH`/`MULTI`/`EXEC` transactions
- Retry logic for handling race conditions (max 3 retries)
- Fail-closed behavior on Redis errors

**Precision:**
- Nanosecond timestamp precision
- 9 decimal places for token storage
- Accurate rate calculations

## Configuration Guidelines

### Choosing Parameters

**Rate (limit):**
- Set based on your service capacity
- Consider downstream dependencies
- Account for burst traffic patterns

**Burst:**
- Allow temporary spikes above the rate
- Typically 2-10x the per-second rate
- Higher burst = more flexible, but higher peak load

### Examples

```go
// API rate limiting
limiter := rate.NewRedisLimiter(client, 100.0, 20) // 100 req/sec, burst 20

// Background job processing  
limiter := rate.NewInMemoryLimiter(5.0, 1) // 5 jobs/sec, no burst

// User-specific limits
limiter := rate.NewRedisLimiter(client, 10.0, 5) // 10 req/sec per user
```

## Error Handling

### In-Memory Limiter
- Always returns a boolean (no errors)
- Zero limit always returns `false`
- Thread-safe operations

### Redis Limiter
- Fail-closed on Redis errors (returns `false`)
- Automatic retries on transaction conflicts
- Graceful degradation when Redis is unavailable

```go
// Redis limiter handles errors gracefully
if limiter.Allow(key) {
    // Definitely allowed
    processRequest()
} else {
    // Either rate limited OR Redis error
    // Fail-closed for safety
    return http.StatusTooManyRequests
}
```

## Performance Considerations

| Aspect | In-Memory | Redis |
|--------|-----------|-------|
| **Latency** | ~10ns | ~1ms |
| **Throughput** | Very High | High |
| **Memory** | O(keys) | O(1) |
| **Persistence** | None | Full |
| **Distribution** | Single process | Multi-process |
| **Scalability** | Vertical only | Horizontal |

## Best Practices

1. **Choose the Right Implementation:**
   - Use in-memory for single-instance applications
   - Use Redis for distributed systems

2. **Key Design:**
   - Use meaningful, consistent key patterns
   - Consider key cardinality (avoid unlimited growth)
   - Examples: `user:{id}`, `api:{endpoint}`, `ip:{address}`

3. **Configuration:**
   - Start with conservative limits
   - Monitor and adjust based on metrics
   - Consider different limits for different user tiers

4. **Error Handling:**
   - Always handle rate limit errors gracefully
   - Provide meaningful error messages to users
   - Consider implementing backoff strategies

5. **Monitoring:**
   - Track rate limit hit rates
   - Monitor Redis performance (for Redis limiter)
   - Alert on unusual patterns

## Testing

The package includes comprehensive tests for both implementations:

```bash
# Run all tests
go test -v .

# Redis tests automatically skip if Redis is unavailable
# Error handling tests work without Redis
```

## Dependencies

- **In-Memory Limiter:** `golang.org/x/time/rate`
- **Redis Limiter:** `github.com/redis/go-redis/v9`
- **Core:** Standard library only

## License

This package is part of the `github.com/unbxd/go-base/v2` library. 