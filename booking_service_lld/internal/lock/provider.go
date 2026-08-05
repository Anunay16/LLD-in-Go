package lock

import (
	"context"
	"time"
)

// Lock represents an acquired distributed/concurrency lock.
type Lock interface {
	Unlock(ctx context.Context) error
}

// LockProvider defines the interface for acquiring and managing locks.
type LockProvider interface {
	// TryAcquire attempts to acquire a lock for a given key without blocking.
	TryAcquire(ctx context.Context, key string, owner string, ttl time.Duration) (Lock, error)

	// Acquire continuously attempts to acquire a lock until context cancels or timeout occurs.
	Acquire(ctx context.Context, key string, owner string, ttl time.Duration) (Lock, error)

	// TryAcquireMulti acquires locks for multiple keys atomically with canonical key sorting to prevent deadlocks.
	// If any key fails to be locked, all previously acquired locks in this call are automatically rolled back.
	TryAcquireMulti(ctx context.Context, keys []string, owner string, ttl time.Duration) ([]Lock, error)

	// IsLockedBy checks if the key is currently locked by the specified owner.
	IsLockedBy(ctx context.Context, key string, owner string) bool
}
