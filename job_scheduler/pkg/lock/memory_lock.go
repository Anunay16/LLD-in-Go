package lock

import (
	"context"
	"sync"
	"time"
)

type lockEntry struct {
	expiresAt time.Time
}

// MemoryLock is an in-memory thread-safe implementation of Locker with TTL.
type MemoryLock struct {
	mu    sync.Mutex
	locks map[string]lockEntry
}

func NewMemoryLock() *MemoryLock {
	return &MemoryLock{
		locks: make(map[string]lockEntry),
	}
}

func (m *MemoryLock) Acquire(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	if entry, exists := m.locks[key]; exists {
		if now.Before(entry.expiresAt) {
			return false, nil // Still held
		}
	}

	m.locks[key] = lockEntry{
		expiresAt: now.Add(ttl),
	}
	return true, nil
}

func (m *MemoryLock) Release(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.locks, key)
	return nil
}
