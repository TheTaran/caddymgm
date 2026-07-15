package main

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func testLoginLimiter(now *time.Time) *loginLimiter {
	return &loginLimiter{
		attempts:        make(map[string]loginAttempt),
		maxFailures:     3,
		failureWindow:   10 * time.Minute,
		lockoutDuration: 5 * time.Minute,
		maxEntries:      4,
		now:             func() time.Time { return *now },
	}
}

func TestLoginLimiterLocksAtThresholdAndExpires(t *testing.T) {
	now := time.Date(2026, time.July, 15, 12, 0, 0, 0, time.UTC)
	limiter := testLoginLimiter(&now)
	key := loginAttemptKeys("admin", "192.0.2.10:1234")[0]

	for attempt := 1; attempt < limiter.maxFailures; attempt++ {
		if _, blocked := limiter.recordFailure(key); blocked {
			t.Fatalf("attempt %d blocked before threshold", attempt)
		}
	}
	if retryAfter, blocked := limiter.recordFailure(key); !blocked || retryAfter != 5*time.Minute {
		t.Fatalf("threshold result = (%s, %t), want (5m, true)", retryAfter, blocked)
	}
	if retryAfter, blocked := limiter.retryAfterAny([]string{key}); !blocked || retryAfter != 5*time.Minute {
		t.Fatalf("active lock result = (%s, %t), want (5m, true)", retryAfter, blocked)
	}

	now = now.Add(5 * time.Minute)
	if retryAfter, blocked := limiter.retryAfterAny([]string{key}); blocked {
		t.Fatalf("expired lock still active for %s", retryAfter)
	}
}

func TestLoginLimiterResetsAfterSuccess(t *testing.T) {
	now := time.Date(2026, time.July, 15, 12, 0, 0, 0, time.UTC)
	limiter := testLoginLimiter(&now)
	key := loginAttemptKeys("admin", "192.0.2.10:1234")[0]

	limiter.recordFailure(key)
	limiter.recordFailure(key)
	limiter.reset(key)
	if _, blocked := limiter.recordFailure(key); blocked {
		t.Fatal("first failure after reset was blocked")
	}
}

func TestLoginLimiterFailureWindowResets(t *testing.T) {
	now := time.Date(2026, time.July, 15, 12, 0, 0, 0, time.UTC)
	limiter := testLoginLimiter(&now)
	key := loginAttemptKeys("admin", "192.0.2.10:1234")[0]

	limiter.recordFailure(key)
	limiter.recordFailure(key)
	now = now.Add(limiter.failureWindow)
	if _, blocked := limiter.recordFailure(key); blocked {
		t.Fatal("failure from an expired window contributed to a lock")
	}
}

func TestLoginLimiterBoundsStoredEntries(t *testing.T) {
	now := time.Date(2026, time.July, 15, 12, 0, 0, 0, time.UTC)
	limiter := testLoginLimiter(&now)

	for i := 0; i < limiter.maxEntries+10; i++ {
		limiter.recordFailure(fmt.Sprintf("key-%d", i))
		now = now.Add(time.Millisecond)
	}
	if got := len(limiter.attempts); got != limiter.maxEntries {
		t.Fatalf("stored entries = %d, want %d", got, limiter.maxEntries)
	}
}

func TestLoginLimiterConcurrentFailures(t *testing.T) {
	now := time.Date(2026, time.July, 15, 12, 0, 0, 0, time.UTC)
	limiter := testLoginLimiter(&now)
	key := loginAttemptKeys("admin", "192.0.2.10:1234")[0]
	var workers sync.WaitGroup

	for i := 0; i < 20; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			limiter.recordFailure(key)
		}()
	}
	workers.Wait()
	if _, blocked := limiter.retryAfterAny([]string{key}); !blocked {
		t.Fatal("concurrent failures did not lock the key")
	}
}

func TestLoginAttemptKeysUseUsernameAndSourceIP(t *testing.T) {
	first := loginAttemptKeys("Admin", "192.0.2.10:1234")
	same := loginAttemptKeys(" admin ", "192.0.2.10:5678")
	if first[0] != same[0] || first[1] != same[1] {
		t.Fatal("username normalization or source port removal changed the keys")
	}
	if first[0] == loginAttemptKeys("other", "192.0.2.10:1234")[0] {
		t.Fatal("different usernames produced the same key")
	}
	if first[1] == loginAttemptKeys("admin", "192.0.2.11:1234")[1] {
		t.Fatal("different source IPs produced the same key")
	}
}

func TestLoginLimiterBlocksUsernameAcrossSourceIPs(t *testing.T) {
	now := time.Date(2026, time.July, 15, 12, 0, 0, 0, time.UTC)
	limiter := testLoginLimiter(&now)

	for i := 0; i < limiter.maxFailures; i++ {
		keys := loginAttemptKeys("admin", fmt.Sprintf("192.0.2.%d:1234", i+1))
		limiter.recordFailure(keys...)
	}
	if _, blocked := limiter.retryAfterAny(loginAttemptKeys("admin", "198.51.100.10:1234")); !blocked {
		t.Fatal("username was not blocked after failures from rotating source IPs")
	}
}
