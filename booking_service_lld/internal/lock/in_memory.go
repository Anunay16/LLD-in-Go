package lock

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

type lockEntry struct {
	owner     string
	expiresAt time.Time
}

// InMemoryLockProvider implements LockProvider using in-memory mutexes.
type InMemoryLockProvider struct {
	mu    sync.Mutex
	locks map[string]*lockEntry
	retry time.Duration
}

// InMemoryLock represents a handle to an acquired in-memory lock.
type InMemoryLock struct {
	key      string
	owner    string
	provider *InMemoryLockProvider
}

// NewInMemoryLockProvider constructs a new in-memory lock provider.
func NewInMemoryLockProvider(retry time.Duration) *InMemoryLockProvider {
	if retry <= 0 {
		retry = 50 * time.Millisecond
	}
	return &InMemoryLockProvider{
		locks: make(map[string]*lockEntry),
		retry: retry,
	}
}

// TryAcquire attempts to acquire a lock for a key. Fails immediately if already locked and not expired.
func (p *InMemoryLockProvider) TryAcquire(ctx context.Context, key string, owner string, ttl time.Duration) (Lock, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	entry, exists := p.locks[key]

	if exists {
		// Check TTL expiration (lazy eviction)
		if now.Before(entry.expiresAt) {
			return nil, ErrLockNotAcquired
		}
	}

	p.locks[key] = &lockEntry{
		owner:     owner,
		expiresAt: now.Add(ttl),
	}

	return &InMemoryLock{
		key:      key,
		owner:    owner,
		provider: p,
	}, nil
}

// Acquire retries acquiring a lock until context timeout/cancellation or success.
func (p *InMemoryLockProvider) Acquire(ctx context.Context, key string, owner string, ttl time.Duration) (Lock, error) {
	ticker := time.NewTicker(p.retry)
	defer ticker.Stop()

	for {
		l, err := p.TryAcquire(ctx, key, owner, ttl)
		if err == nil {
			return l, nil
		}

		if err != ErrLockNotAcquired {
			return nil, err
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
		}
	}
}

// TryAcquireMulti acquires locks for multiple keys.
// Crucial: It sorts keys lexicographically to enforce global locking order and prevent DEADLOCKS.
// If any single lock cannot be acquired, all previously acquired locks are released (atomic rollback).
func (p *InMemoryLockProvider) TryAcquireMulti(ctx context.Context, keys []string, owner string, ttl time.Duration) ([]Lock, error) {
	if len(keys) == 0 {
		return nil, nil
	}

	// 1. Canonical Sorting to prevent deadlock across concurrent requests
	sortedKeys := make([]string, len(keys))
	copy(sortedKeys, keys)
	sort.Strings(sortedKeys)
	sortedKeys = deduplicate(sortedKeys)

	acquired := make([]Lock, 0, len(sortedKeys))

	// 2. Acquire each lock sequentially according to canonical order
	for _, key := range sortedKeys {
		l, err := p.TryAcquire(ctx, key, owner, ttl)
		if err != nil {
			// Rollback already acquired locks in reverse order
			for i := len(acquired) - 1; i >= 0; i-- {
				_ = acquired[i].Unlock(ctx)
			}
			return nil, fmt.Errorf("%w: failed on resource '%s'", ErrLockNotAcquired, key)
		}
		acquired = append(acquired, l)
	}

	return acquired, nil
}

func (l *InMemoryLock) Unlock(ctx context.Context) error {
	return l.provider.release(l.key, l.owner)
}

func (p *InMemoryLockProvider) release(key string, owner string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	entry, exists := p.locks[key]
	if !exists {
		return ErrLockNotOwnedOrReleased
	}

	// Safety check: ensure only the owner who acquired the lock can release it
	if entry.owner != owner {
		return ErrLockNotOwnedOrReleased
	}

	delete(p.locks, key)
	return nil
}

// IsLockedBy checks if a given key is held by owner and has not expired.
func (p *InMemoryLockProvider) IsLockedBy(ctx context.Context, key string, owner string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	entry, exists := p.locks[key]
	if !exists {
		return false
	}

	if time.Now().After(entry.expiresAt) {
		return false
	}

	return entry.owner == owner
}

func deduplicate(sorted []string) []string {
	if len(sorted) <= 1 {
		return sorted
	}
	j := 0
	for i := 1; i < len(sorted); i++ {
		if sorted[i] != sorted[j] {
			j++
			sorted[j] = sorted[i]
		}
	}
	return sorted[:j+1]
}
