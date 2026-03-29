package main

import (
	"context"
	"fmt"
	"time"

	"lockingsystem/internal/lock"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func main() {
	ctx := context.Background()

	// 1️⃣ Create Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// 2️⃣ Create Lock Provider
	lockProvider := lock.NewRedisLockProvider(rdb, 50*time.Millisecond)

	// 3️⃣ Create context with timeout (IMPORTANT)
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// 4️⃣ Acquire lock
	owner1 := uuid.NewString()
	l, err := lockProvider.Acquire(ctxWithTimeout, "order:123", owner1, 10*time.Second)
	if err != nil {
		fmt.Println("failed to acquire lock:", err)
		return
	}
	defer l.Unlock(context.Background())

	fmt.Println("Lock acquired!")

	// 5️⃣ Critical section
	doWork()

	fmt.Println("Work done, lock released")
}

func doWork() {
	fmt.Println("processing...")
	time.Sleep(2 * time.Second)
}
