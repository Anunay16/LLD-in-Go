package lock

import (
	"context"
	"sync"
	"time"
)

type InMemoryLockProvider struct {
	mu    sync.Mutex
	locks map[string]*lockEntry
	retry time.Duration
}

type lockEntry struct {
	owner     string
	mu        sync.Mutex
	expiresAt time.Time
}

type InMemoryLock struct {
	key      string
	owner    string
	entry    *lockEntry
	provider *InMemoryLockProvider
}

func NewInMemoryLockProvider(retry time.Duration) *InMemoryLockProvider {
	if retry <= 0 {
		retry = 50 * time.Millisecond
	}
	return &InMemoryLockProvider{
		locks: make(map[string]*lockEntry),
		retry: retry,
	}
}

func (p *InMemoryLockProvider) TryAcquire(ctx context.Context, key string, owner string, ttl time.Duration) (Lock, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	entry, exists := p.locks[key]

	if exists && now.Before(entry.expiresAt) {
		return nil, ErrLockNotAcquired
	}

	entry = &lockEntry{
		owner:     owner,
		expiresAt: now.Add(ttl),
	}

	entry.mu.Lock()
	p.locks[key] = entry

	return &InMemoryLock{
		key:      key,
		owner:    owner,
		provider: p,
		entry:    entry,
	}, nil
}

func (p *InMemoryLockProvider) Acquire(ctx context.Context, key string, owner string, ttl time.Duration) (Lock, error) {
	ticker := time.NewTicker(p.retry)
	defer ticker.Stop()

	for {
		lock, err := p.TryAcquire(ctx, key, owner, ttl)
		if err == nil {
			return lock, nil
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

func (l *InMemoryLock) Unlock(ctx context.Context) error {
	p := l.provider

	p.mu.Lock()
	defer p.mu.Unlock()

	entry, exists := p.locks[l.key]
	if !exists {
		return ErrLockNotOwnedOrReleased
	}

	// ownership check
	if entry != l.entry || entry.owner != l.owner {
		return ErrLockNotOwnedOrReleased
	}

	entry.mu.Unlock()
	delete(p.locks, l.key)

	return nil
}

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
