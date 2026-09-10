# GitLab Go Code Review Interview Preparation Guide

This document captures real-world Go merge request (MR) scenarios focusing on concurrency control, channels, `sync.WaitGroup`, race conditions, memory leaks, and design patterns.

---

# Table of Contents
1. [MR #1: Async Event Dispatcher / Worker Pool](#mr-1-async-event-dispatcher--worker-pool)
2. [MR #2: Cache with Read-Through & Design Patterns](#mr-2-cache-with-read-through--design-patterns)
3. [MR #3: Pub/Sub Broker with Graceful Shutdown](#mr-3-pubsub-broker-with-graceful-shutdown)

---

## MR #1: Async Event Dispatcher / Worker Pool

### 1. MR Description
> *"We need to process incoming webhook events asynchronously in batches. This MR adds a worker pool that accepts tasks, processes them with worker goroutines, and waits for all of them to complete before returning results."*

### 2. The Code (Under Review)
```go
package worker

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Event struct {
	ID      string
	Payload string
}

type Result struct {
	EventID string
	Err     error
}

// ProcessEvents processes a slice of events using a pool of concurrent workers.
func ProcessEvents(ctx context.Context, events []Event, concurrency int) ([]Result, error) {
	var wg sync.WaitGroup
	eventCh := make(chan Event)
	resultCh := make(chan Result)
	results := make([]Result, 0, len(events))

	// 1. Start worker pool
	for i := 0; i < concurrency; i++ {
		go func(workerID int, wg sync.WaitGroup) {
			defer wg.Done()
			for event := range eventCh {
				res := handleEvent(ctx, event)
				resultCh <- res
			}
		}(i, wg)
	}

	// 2. Feed events to the worker pool
	go func() {
		for _, event := range events {
			eventCh <- event
		}
		close(eventCh)
	}()

	// 3. Collect results
	for res := range resultCh {
		results = append(results, res)
	}

	wg.Wait()
	return results, nil
}

func handleEvent(ctx context.Context, e Event) Result {
	select {
	case <-ctx.Done():
		return Result{EventID: e.ID, Err: ctx.Err()}
	case <-time.After(10 * time.Millisecond):
		return Result{EventID: e.ID, Err: nil}
	}
}
```

### 3. Why It Is Bad (Flaws & Failure Modes)
1. **`sync.WaitGroup` Copied by Value:**
   - `sync.WaitGroup` contains internal state (an atomic 64-bit value and semaphore). Passing it by value copies this state. Calling `Done()` on the copy has zero effect on the caller's `wg`.
   - `go vet` flags this with `copylocks`.
2. **Missing `wg.Add()`:**
   - Even if passed by reference, `wg.Add` was never called. `wg.Wait()` sees `0` and exits without waiting.
3. **Deadlock on `resultCh` Iteration:**
   - `for res := range resultCh` blocks forever waiting for elements or until `resultCh` is closed.
   - `resultCh` is never closed, and `wg.Wait()` sits *after* the collection loop, so `wg.Wait()` is never reached. Program deadlocks immediately.
4. **Goroutine Leak in Producer on Context Cancellation:**
   - The producer goroutine loops over `events` sending to unbuffered `eventCh`. If workers exit due to `ctx.Done()`, no one reads `eventCh`. The producer goroutine hangs forever on `eventCh <- event`.
5. **No Concurrency Bounds Validation:**
   - If `concurrency <= 0`, no workers spawn, and both producer and consumer deadlock.

### 4. Ideal Review Comments to Leave on the MR
- **On `go func(workerID int, wg sync.WaitGroup)`:**
  > *"Avoid copying `sync.WaitGroup` by value. A `WaitGroup` must be passed by pointer (`*sync.WaitGroup`) or captured from the outer lexical scope. Also, make sure to call `wg.Add(1)` in the parent goroutine before spawning each worker (or `wg.Add(concurrency)` prior to the loop)."*
- **On `for res := range resultCh` and `wg.Wait()`:**
  > *"This causes an immediate deadlock. `range resultCh` will block indefinitely because `resultCh` is never closed. Because `wg.Wait()` is placed after this loop, it is never reached. To fix this, spawn a dedicated coordinator goroutine: `go func() { wg.Wait(); close(resultCh) }()` before reading from `resultCh`."*
- **On Producer Goroutine `eventCh <- event`:**
  > *"If the context is cancelled early and workers exit, this producer goroutine will leak because it blocks indefinitely on `eventCh <- event`. Please add a `select` with `case <-ctx.Done(): return`."*
- **On function entry:**
  > *"Please guard against `concurrency <= 0` (either default to 1 or return an error) to prevent deadlocks when no workers are spawned."*

### 5. Recommended Solution
```go
func ProcessEvents(ctx context.Context, events []Event, concurrency int) ([]Result, error) {
	if concurrency <= 0 {
		concurrency = 1
	}

	eventCh := make(chan Event)
	resultCh := make(chan Result)
	results := make([]Result, 0, len(events))

	var wg sync.WaitGroup

	// 1. Start worker pool
	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func(workerID int) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case event, ok := <-eventCh:
					if !ok {
						return
					}
					result := handleEvent(ctx, event)
					select {
					case resultCh <- result:
					case <-ctx.Done():
						return
					}
				}
			}
		}(i)
	}

	// 2. Feed events safely with context cancellation
	go func() {
		defer close(eventCh)
		for _, event := range events {
			select {
			case <-ctx.Done():
				return
			case eventCh <- event:
			}
		}
	}()

	// 3. Close result channel once all workers finish
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// 4. Drain results
	for res := range resultCh {
		results = append(results, res)
	}

	return results, ctx.Err()
}
```

---

## MR #2: Cache with Read-Through & Design Patterns

### 1. MR Description
> *"This MR implements a thread-safe in-memory cache manager using a Singleton pattern. It features an Expiring Cache with read-through loading to prevent cache stampede (multiple concurrent callers hitting the database for the same key)."*

### 2. The Code (Under Review)
```go
package cache

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Item struct {
	Value      any
	Expiration time.Time
}

type CacheManager struct {
	mu    sync.RWMutex
	store map[string]Item
}

var instance *CacheManager
var mu sync.Mutex

// GetInstance returns the singleton CacheManager
func GetInstance() *CacheManager {
	if instance == nil {
		mu.Lock()
		defer mu.Unlock()
		if instance == nil {
			instance = &CacheManager{
				store: make(map[string]Item),
			}
		}
	}
	return instance
}

// GetOrFetch retrieves an item, or calls fetcher if missing or expired.
func (c *CacheManager) GetOrFetch(ctx context.Context, key string, ttl time.Duration, fetcher func(context.Context) (any, error)) (any, error) {
	c.mu.RLock()
	item, found := c.store[key]
	c.mu.RUnlock()

	if found && time.Now().Before(item.Expiration) {
		return item.Value, nil
	}

	// Cache miss or expired: fetch new value
	c.mu.Lock()
	defer c.mu.Unlock()

	// Fetch data
	val, err := fetcher(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetcher failed: %w", err)
	}

	c.store[key] = Item{
		Value:      val,
		Expiration: time.Now().Add(ttl),
	}

	return val, nil
}
```

### 3. Why It Is Bad (Flaws & Failure Modes)
1. **Global Singleton Anti-Pattern & Testability:**
   - Makes parallel unit tests (`t.Parallel()`) impossible without race conditions or side effects polluting across tests.
   - Go strongly favors explicit constructors `New(...)` and dependency injection via interfaces.
2. **Broken Double-Checked Locking:**
   - The initial `if instance == nil` is an unsynchronized read. Writing to `instance` in another thread without memory barrier/atomic read causes a data race under `-race`.
   - Go provides `sync.Once` for exactly this purpose.
3. **Severe Lock Contention on `fetcher(ctx)`:**
   - Calling a potentially slow network/DB `fetcher` while holding the global write lock (`c.mu.Lock()`) blocks all reads and writes across **every single key** in the entire cache.
4. **Does NOT Prevent Cache Stampede / Missing Double-Check:**
   - If two requests check key `"user:1"` on a miss, both pass the `RLock`.
   - Thread 1 gets `Lock()`, calls `fetcher()`, updates cache, unlocks.
   - Thread 2 gets `Lock()`, and **immediately executes `fetcher()` again** without checking `c.store[key]`.
   - If you move `fetcher()` outside the lock, all concurrent requests hit the DB at once (thundering herd).

### 4. Ideal Review Comments to Leave on the MR
- **On Singleton / `GetInstance`:**
  > *"Global singletons hinder testability and prevent running tests with `t.Parallel()`. Please consider exporting a constructor like `New() *Cache` and accepting it via dependency injection. If a singleton is strictly needed, use `sync.Once` instead of custom double-checked locking to avoid unsynchronized reads and data races."*
- **On Lock Contention during `fetcher(ctx)`:**
  > *"Executing `fetcher(ctx)` while holding `c.mu.Lock()` creates a major bottleneck: any slow I/O call will block read and write access for the entire cache across all keys."*
- **On Cache Stampede / Redundant Fetches:**
  > *"Inside `GetOrFetch`, once `c.mu.Lock()` is acquired, we do not re-check if another goroutine has already populated `key`. Additionally, to truly prevent cache stampedes without locking the entire cache, consider using `golang.org/x/sync/singleflight` to de-duplicate concurrent requests per key."*

### 5. Recommended Solution
```go
package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

type Item struct {
	Value      any
	Expiration time.Time
}

type Cache struct {
	mu    sync.RWMutex
	store map[string]Item
	sfg   singleflight.Group
}

// New returns a new Cache instance suitable for dependency injection
func New() *Cache {
	return &Cache{
		store: make(map[string]Item),
	}
}

func (c *Cache) GetOrFetch(ctx context.Context, key string, ttl time.Duration, fetcher func(context.Context) (any, error)) (any, error) {
	// 1. Fast path: RLock check
	c.mu.RLock()
	item, found := c.store[key]
	c.mu.RUnlock()

	if found && time.Now().Before(item.Expiration) {
		return item.Value, nil
	}

	// 2. Singleflight: Only 1 fetcher runs concurrently per key across all goroutines
	v, err, _ := c.sfg.Do(key, func() (any, error) {
		// Double check under lock in case another singleflight call just finished
		c.mu.RLock()
		item, found := c.store[key]
		c.mu.RUnlock()
		if found && time.Now().Before(item.Expiration) {
			return item.Value, nil
		}

		// Perform fetch without holding any cache-wide lock
		val, err := fetcher(ctx)
		if err != nil {
			return nil, err
		}

		c.mu.Lock()
		c.store[key] = Item{
			Value:      val,
			Expiration: time.Now().Add(ttl),
		}
		c.mu.Unlock()

		return val, nil
	})

	if err != nil {
		return nil, fmt.Errorf("fetch failed: %w", err)
	}

	return v, nil
}
```

---

## MR #3: Pub/Sub Broker with Graceful Shutdown

### 1. MR Description
> *"Implements an in-memory PubSub topic broker. Subscribers register their own channels to receive broadcasted messages. Added a `Close()` method to gracefully shut down the broker and all subscriber channels."*

### 2. The Code (Under Review)
```go
package pubsub

import (
	"errors"
	"sync"
)

type Broker struct {
	mu          sync.RWMutex
	subscribers map[chan string]struct{}
	closed      bool
}

func NewBroker() *Broker {
	return &Broker{
		subscribers: make(map[chan string]struct{}),
	}
}

// Subscribe returns a channel that receives published messages.
func (b *Broker) Subscribe() <-chan string {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return nil
	}

	ch := make(chan string)
	b.subscribers[ch] = struct{}{}
	return ch
}

// Publish broadcasts a message to all active subscribers.
func (b *Broker) Publish(msg string) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed {
		return errors.New("broker is closed")
	}

	for ch := range b.subscribers {
		ch <- msg
	}

	return nil
}

// Close gracefully closes the broker and closes all subscriber channels.
func (b *Broker) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return
	}

	b.closed = true

	for ch := range b.subscribers {
		close(ch)
		delete(b.subscribers, ch)
	}
}
```

### 3. Why It Is Bad (Flaws & Failure Modes)
1. **Unbuffered Channels Block the Entire Broker:**
   - Channels are unbuffered (`make(chan string)`).
   - If even one subscriber takes time to read, `ch <- msg` blocks synchronously.
   - Because `Publish` holds `b.mu.RLock()`, while it is blocked:
     - New subscribers cannot join (`Subscribe` needs `Lock()`).
     - Broker cannot shut down (`Close` needs `Lock()`).
2. **RWMutex Starvation & System-Wide Deadlock:**
   - If Goroutine 1 is blocked on `ch <- msg` holding `RLock()`, and Goroutine 2 calls `Close()` (waiting for `Lock()`), any subsequent call to `Publish()` from Goroutine 3 will also block because Go's `RWMutex` prioritizes pending writers. Every caller freezes.
3. **No Unsubscribe Mechanism:**
   - If a subscriber finishes its work and exits, its channel stays in `b.subscribers`. Future `Publish()` calls will hang forever trying to send to a goroutine that no longer exists.
4. **Risk of Panic on Close vs Unsubscribe:**
   - In Go: *Never close a channel from the receiver side, and never send on a closed channel.*
   - If an ad-hoc unsubscribe closes the channel while `Publish()` is executing `ch <- msg`, the program panics with `panic: send on closed channel`.

### 4. Ideal Review Comments to Leave on the MR
- **On `ch <- msg` inside `Publish` under `RLock`:**
  > *"Holding `b.mu.RLock()` while writing to unbuffered channels introduces severe head-of-line blocking and deadlock risks. If one subscriber is slow or unresponsive, `Publish` hangs holding the read lock. This blocks `Close()` and blocks all subsequent `Publish` operations due to Go `RWMutex` writer-priority semantics."*
- **On Snapshotting Subscribers:**
  > *"To decouple locking from message dispatching, copy the subscriber list under lock, release the lock immediately, and then broadcast to the channels."*
- **On Unbuffered vs Non-blocking / Buffered Dispatch:**
  > *"Allow subscribers to specify a buffer size or use a non-blocking `select` with `default:` (or context timeout) so that slow subscribers don't hang the entire publishing system."*
- **On Unsubscribe Support:**
  > *"We need an `Unsubscribe` mechanism. A clean Go idiom is returning a cleanup closure from `Subscribe`: `(<-chan string, func())`."*

### 5. Recommended Solution
```go
package pubsub

import (
	"context"
	"errors"
	"sync"
)

var (
	ErrBrokerClosed = errors.New("broker is closed")
)

type Broker struct {
	mu          sync.RWMutex
	subscribers map[chan string]struct{}
	closed      bool
}

func NewBroker() *Broker {
	return &Broker{
		subscribers: make(map[chan string]struct{}),
	}
}

// Subscribe registers a subscriber with a configurable buffer and returns an unsubscribe func.
func (b *Broker) Subscribe(bufSize int) (<-chan string, func()) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return nil, func() {}
	}

	ch := make(chan string, bufSize)
	b.subscribers[ch] = struct{}{}

	unsubscribe := func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		if _, ok := b.subscribers[ch]; ok {
			delete(b.subscribers, ch)
			close(ch)
		}
	}

	return ch, unsubscribe
}

// Publish broadcasts a message without holding locks during channel send.
func (b *Broker) Publish(ctx context.Context, msg string) error {
	b.mu.RLock()
	if b.closed {
		b.mu.RUnlock()
		return ErrBrokerClosed
	}

	// 1. Snapshot subscribers under read lock, then release immediately
	subs := make([]chan string, 0, len(b.subscribers))
	for ch := range b.subscribers {
		subs = append(subs, ch)
	}
	b.mu.RUnlock()

	// 2. Broadcast without holding lock
	for _, ch := range subs {
		select {
		case ch <- msg:
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Non-blocking drop or log warning to prevent slow subscriber bottleneck
		}
	}

	return nil
}

// Close gracefully stops the broker and closes all remaining channels.
func (b *Broker) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return
	}
	b.closed = true

	for ch := range b.subscribers {
		close(ch)
	}
	b.subscribers = nil
}
```
