package rate

import (
	"testing"
	"time"
)

func TestNewInMemoryLimiter(t *testing.T) {
	limiter := NewInMemoryLimiter(1, 2)
	if limiter == nil {
		t.Fatal("Expected non-nil limiter")
	}
}

func TestAllow_BurstAndRate(t *testing.T) {
	tests := []struct {
		name      string
		limit     float64
		burst     int
		requests  int
		sleep     time.Duration
		wantAllow []bool
	}{
		{
			name:      "burst=2, limit=2, 3 requests, no wait",
			limit:     2,
			burst:     2,
			requests:  3,
			sleep:     0,
			wantAllow: []bool{true, true, false},
		},
		{
			name:      "burst=1, limit=1, 2 requests, no wait",
			limit:     1,
			burst:     1,
			requests:  2,
			sleep:     0,
			wantAllow: []bool{true, false},
		},
		{
			name:      "burst=5, limit=2, 6 requests, no wait",
			limit:     2,
			burst:     5,
			requests:  6,
			sleep:     0,
			wantAllow: []bool{true, true, true, true, true, false},
		},
		{
			name:      "burst=3, limit=1, 3 requests, wait for recovery",
			limit:     1,
			burst:     3,
			requests:  3,
			sleep:     1100 * time.Millisecond,
			wantAllow: []bool{true, true, true}, // after wait, should allow again
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter := NewInMemoryLimiter(tt.limit, tt.burst)
			key := Key("user1")
			for i := 0; i < len(tt.wantAllow); i++ {
				allowed := limiter.Allow(key)
				if allowed != tt.wantAllow[i] {
					t.Errorf("Request %d: got %v, want %v", i+1, allowed, tt.wantAllow[i])
				}
			}
			if tt.sleep > 0 {
				time.Sleep(tt.sleep)
				allowed := limiter.Allow(key)
				if !allowed {
					t.Error("Expected event to be allowed after waiting for rate recovery")
				}
			}
		})
	}
}

func TestAllow_MultipleKeysIndependence(t *testing.T) {
	limiter := NewInMemoryLimiter(2, 2)
	key1 := Key("user1")
	key2 := Key("user2")

	// Both should allow burst
	if !limiter.Allow(key1) {
		t.Error("Expected key1 first event to be allowed")
	}
	if !limiter.Allow(key2) {
		t.Error("Expected key2 first event to be allowed")
	}
	if !limiter.Allow(key1) {
		t.Error("Expected key1 second event to be allowed (burst)")
	}
	if !limiter.Allow(key2) {
		t.Error("Expected key2 second event to be allowed (burst)")
	}
	// Both should now be rate limited
	if limiter.Allow(key1) {
		t.Error("Expected key1 to be rate limited after burst")
	}
	if limiter.Allow(key2) {
		t.Error("Expected key2 to be rate limited after burst")
	}
}

func TestAllow_ZeroBurst(t *testing.T) {
	limiter := NewInMemoryLimiter(1, 0)
	key := Key("user1")
	if limiter.Allow(key) {
		t.Error("Expected event to be rate limited with zero burst")
	}
}

func TestAllow_ZeroLimit(t *testing.T) {
	limiter := NewInMemoryLimiter(0, 1)
	key := Key("user1")
	if limiter.Allow(key) {
		t.Error("Expected event to be rate limited with zero limit")
	}
}

func TestAllow_HighBurst(t *testing.T) {
	limiter := NewInMemoryLimiter(1, 10)
	key := Key("user1")
	for i := 0; i < 10; i++ {
		if !limiter.Allow(key) {
			t.Errorf("Expected event %d to be allowed in high burst", i+1)
		}
	}
	if limiter.Allow(key) {
		t.Error("Expected event to be rate limited after high burst consumed")
	}
}
