package lock

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// RedisLock simulates a distributed lock using Redis semantics (SET key token NX PX ttl).
// In a real production deployment, this wraps go-redis with Redlock or SET NX PX + Lua script unlock.
type RedisLock struct {
	mu     sync.Mutex
	store  map[string]redisLockItem
	client string // e.g. "redis://localhost:6379"
}

type redisLockItem struct {
	token     string
	expiresAt time.Time
}

func NewRedisLock(address string) *RedisLock {
	return &RedisLock{
		client: address,
		store:  make(map[string]redisLockItem),
	}
}

func (r *RedisLock) Acquire(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	if item, exists := r.store[key]; exists && now.Before(item.expiresAt) {
		return false, nil
	}

	r.store[key] = redisLockItem{
		token:     fmt.Sprintf("token-%d", now.UnixNano()),
		expiresAt: now.Add(ttl),
	}
	return true, nil
}

func (r *RedisLock) Release(ctx context.Context, key string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.store, key)
	return nil
}
