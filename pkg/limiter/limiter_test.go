package limiter_test

import (
	"context"
	"testing"
	"time"

	"github.com/hanjo-erp/hanjo/pkg/limiter"
)

func TestMemoryRateLimiter_LockoutAfterMaxAttempts(t *testing.T) {
	ctx := context.Background()
	rl := limiter.NewMemoryRateLimiter()
	key := "user:123"

	// Attempts 1 to 4 should not lock
	for i := 1; i <= 4; i++ {
		attempts, remaining, lockedUntil, err := rl.RecordFailedAttempt(ctx, key, 5, 15*time.Minute)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if attempts != i {
			t.Errorf("expected %d attempts, got %d", i, attempts)
		}
		if remaining != 5-i {
			t.Errorf("expected %d remaining, got %d", 5-i, remaining)
		}
		if lockedUntil != nil {
			t.Errorf("expected not locked, got locked until: %v", lockedUntil)
		}
	}

	// 5th attempt should lock
	attempts, remaining, lockedUntil, err := rl.RecordFailedAttempt(ctx, key, 5, 15*time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if attempts != 5 {
		t.Errorf("expected 5 attempts, got %d", attempts)
	}
	if remaining != 0 {
		t.Errorf("expected 0 remaining, got %d", remaining)
	}
	if lockedUntil == nil {
		t.Errorf("expected lockedUntil to be set")
	}

	// CheckLockout should return true
	locked, until, err := rl.CheckLockout(ctx, key)
	if err != nil || !locked || until == nil {
		t.Errorf("expected locked=true, got %v", locked)
	}

	// Reset should clear lockout
	if err := rl.Reset(ctx, key); err != nil {
		t.Fatalf("unexpected reset error: %v", err)
	}

	locked, _, _ = rl.CheckLockout(ctx, key)
	if locked {
		t.Errorf("expected not locked after reset")
	}
}
