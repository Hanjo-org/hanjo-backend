package limiter

import (
	"context"
	"sync"
	"time"
)

// RateLimiter manages failed attempt counters and lockout periods.
type RateLimiter interface {
	RecordFailedAttempt(ctx context.Context, key string, maxAttempts int, lockDuration time.Duration) (attempts int, remaining int, lockedUntil *time.Time, err error)
	CheckLockout(ctx context.Context, key string) (isLocked bool, lockedUntil *time.Time, err error)
	Reset(ctx context.Context, key string) error
}

type attemptRecord struct {
	count       int
	lockedUntil *time.Time
	updatedAt   time.Time
}

// MemoryRateLimiter is an in-memory thread-safe rate limiter.
type MemoryRateLimiter struct {
	mu      sync.RWMutex
	records map[string]*attemptRecord
}

// NewMemoryRateLimiter creates a new MemoryRateLimiter instance.
func NewMemoryRateLimiter() *MemoryRateLimiter {
	return &MemoryRateLimiter{
		records: make(map[string]*attemptRecord),
	}
}

// RecordFailedAttempt increments the failure count and locks if maxAttempts is reached.
func (m *MemoryRateLimiter) RecordFailedAttempt(
	_ context.Context,
	key string,
	maxAttempts int,
	lockDuration time.Duration,
) (int, int, *time.Time, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	rec, exists := m.records[key]
	if !exists {
		rec = &attemptRecord{count: 0}
		m.records[key] = rec
	}

	// If previously locked but lock expired, reset
	if rec.lockedUntil != nil && now.After(*rec.lockedUntil) {
		rec.count = 0
		rec.lockedUntil = nil
	}

	rec.count++
	rec.updatedAt = now

	remaining := maxAttempts - rec.count
	if remaining < 0 {
		remaining = 0
	}

	if rec.count >= maxAttempts {
		until := now.Add(lockDuration)
		rec.lockedUntil = &until
		return rec.count, 0, rec.lockedUntil, nil
	}

	return rec.count, remaining, nil, nil
}

// CheckLockout checks if a key is currently locked out.
func (m *MemoryRateLimiter) CheckLockout(_ context.Context, key string) (bool, *time.Time, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	rec, exists := m.records[key]
	if !exists || rec.lockedUntil == nil {
		return false, nil, nil
	}

	now := time.Now()
	if now.Before(*rec.lockedUntil) {
		return true, rec.lockedUntil, nil
	}

	return false, nil, nil
}

// Reset clears the record for a key upon successful login.
func (m *MemoryRateLimiter) Reset(_ context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.records, key)
	return nil
}
