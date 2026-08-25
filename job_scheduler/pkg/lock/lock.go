package lock

import (
	"context"
	"time"
)

// Locker provides distributed/concurrency locking primitives to prevent concurrent executions of the same job.
type Locker interface {
	Acquire(ctx context.Context, key string, ttl time.Duration) (bool, error)
	Release(ctx context.Context, key string) error
}
