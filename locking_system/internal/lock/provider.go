package lock

import (
	"context"
	"time"
)

type Lock interface {
	Unlock(ctx context.Context) error
}

type LockProvider interface {
	TryAcquire(ctx context.Context, key string, owner string, ttl time.Duration) (Lock, error)
	Acquire(ctx context.Context, key string, owner string, ttl time.Duration) (Lock, error)
	IsLockedBy(ctx context.Context, key string, owner string) bool
}
