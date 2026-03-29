package lock

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisLockProvider struct {
	client *redis.Client
	retry  time.Duration
}

type RedisLock struct {
	key    string
	value  string
	client *redis.Client
}

func NewRedisLockProvider(client *redis.Client, retry time.Duration) *RedisLockProvider {
	if retry <= 0 {
		retry = 50 * time.Millisecond
	}
	return &RedisLockProvider{client: client, retry: retry}
}

func (p *RedisLockProvider) Acquire(ctx context.Context, key string, owner string, ttl time.Duration) (Lock, error) {
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

func (p *RedisLockProvider) TryAcquire(ctx context.Context, key string, owner string, ttl time.Duration) (Lock, error) {
	val := owner

	res, err := p.client.SetArgs(ctx, key, val, redis.SetArgs{
		TTL:  ttl,
		Mode: "NX",
	}).Result()
	if err != nil {
		return nil, err
	}
	if res != "OK" {
		return nil, ErrLockNotAcquired
	}

	return &RedisLock{
		key:    key,
		value:  val,
		client: p.client,
	}, nil
}

var unlockScript = redis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("DEL", KEYS[1])
else
	return 0
end
`)

func (rl *RedisLock) Unlock(ctx context.Context) error {
	res, err := unlockScript.Run(ctx, rl.client, []string{rl.key}, rl.value).Result()
	if err != nil {
		return err
	}

	val, ok := res.(int64)
	if !ok || val == 0 {
		return ErrLockNotOwnedOrReleased
	}

	return nil
}

func (p *RedisLockProvider) IsLockedBy(ctx context.Context, key string, owner string) bool {
	val, err := p.client.Get(ctx, key).Result()
	if err != nil {
		return false
	}

	return val == owner
}
