# GitLab Go Code Review Interview Preparation Guide

This document captures real-world Go merge request (MR) scenarios focusing on concurrency control, channels, `sync.WaitGroup`, race conditions, memory leaks, and design patterns.

---

# Table of Contents
1. [MR #1: Async Event Dispatcher / Worker Pool](#mr-1-async-event-dispatcher--worker-pool)
2. [MR #2: Cache with Read-Through & Design Patterns](#mr-2-cache-with-read-through--design-patterns)
3. [MR #3: Pub/Sub Broker with Graceful Shutdown](#mr-3-pubsub-broker-with-graceful-shutdown)
4. [MR #4: Batch File Processor Pipeline](#mr-4-batch-file-processor-pipeline)
5. [MR #5: Distributed Rate Limiter & Token Bucket](#mr-5-distributed-rate-limiter--token-bucket)
6. [MR #6: Event Listener & Notifier Registry](#mr-6-event-listener--notifier-registry)

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


---

## MR #4: Batch File Processor Pipeline

### 1. MR Description
> *"We process large files in stages using a pipeline pattern: `Reader -> Transformer -> Uploader`. This MR implements the pipeline with cancellation support so that if any stage errors out, all stages terminate immediately without leaking goroutines."*

### 2. The Code (Under Review)
```go
package pipeline

import (
	"context"
	"fmt"
	"sync"
)

type Item struct {
	ID   int
	Data string
}

// RunPipeline runs the 3-stage pipeline concurrently
func RunPipeline(ctx context.Context, totalItems int) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	readCh := make(chan Item)
	transformCh := make(chan Item)
	errCh := make(chan error, 1)

	var wg sync.WaitGroup

	// Stage 1: Producer / Reader
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(readCh)
		for i := 0; i < totalItems; i++ {
			select {
			case <-ctx.Done():
				return
			case readCh <- Item{ID: i, Data: fmt.Sprintf("raw-%d", i)}:
			}
		}
	}()

	// Stage 2: Transformer
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(transformCh)
		for item := range readCh {
			if item.ID == 13 { // Simulated error condition
				select {
				case errCh <- fmt.Errorf("bad item id %d", item.ID):
				default:
				}
				cancel()
				return
			}
			transformCh <- Item{ID: item.ID, Data: item.Data + "-transformed"}
		}
	}()

	// Stage 3: Uploader (Consumer)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for item := range transformCh {
			_ = item // simulated upload
		}
	}()

	wg.Wait()

	select {
	case err := <-errCh:
		return err
	default:
		return nil
	}
}
```

### 3. Why It Is Bad (Flaws & Failure Modes)
1. **Unprotected Downstream Send (`transformCh <- ...`) Causes Goroutine Leak / Deadlock:**
   - In Stage 2, `transformCh <- Item{...}` is an unbuffered blocking send not wrapped in a `select` with `case <-ctx.Done()`.
   - If downstream (Stage 3) terminates early or stops reading, Stage 2 blocks forever on sending. It will never read the next item and will never respond to context cancellation.
2. **Reading with `for item := range readCh` Ignores Cancellation When Idle:**
   - In Stage 2, `for item := range readCh` only checks for channel close or new items. If Stage 1 pauses or is slow, Stage 2 blocks indefinitely on receiving without reacting to `<-ctx.Done()`.
3. **Exit Race Condition (Stage 1 Leaks):**
   - When Stage 2 exits on error (`item.ID == 13`), it terminates and no longer reads `readCh`.
   - If Stage 1 is already blocked attempting to send on `readCh`, it will never unblock. Because Go select is pseudo-random when multiple cases are ready, if Stage 1 is blocked waiting for a reader, it will hang.
   - Stage 1 never calls `wg.Done()`, causing `wg.Wait()` to deadlock the entire program.
4. **Stage 3 Consumer Does Not Check Context:**
   - Stage 3 continues draining and executing work even after context is cancelled.
5. **Manual Synchronization Boilerplate:**
   - Coordinating `sync.WaitGroup`, `errCh`, and manual `cancel()` calls is error-prone and easily replaced by standard Go tooling.

### 4. Ideal Review Comments to Leave on the MR
- **On Stage 2 `transformCh <- Item{...}`:**
  > *"This send is not guarded by context cancellation. If Stage 3 exits or halts, Stage 2 will block indefinitely here, causing a goroutine leak. Every channel send in a pipeline must be wrapped in a `select` with `case <-ctx.Done(): return ctx.Err()`."*
- **On Stage 2 `for item := range readCh`:**
  > *"Using `for item := range readCh` prevents Stage 2 from reacting to context cancellation while waiting for new items. Use an explicit `select` between `case <-ctx.Done():` and `case item, ok := <-readCh:`."*
- **On Stage 3 Execution:**
  > *"Stage 3 does not check `ctx.Done()`. If upstream stages cancel or error out, Stage 3 should immediately abort processing pending items."*
- **Design Recommendation (`errgroup`):**
  > *"We are manually coordinating `sync.WaitGroup`, `errCh`, and manual cancellation. Consider adopting `golang.org/x/sync/errgroup` (`errgroup.WithContext`). It handles goroutine lifecycle, automatically cancels the shared context on the first non-nil error, and captures that error cleanly."*

### 5. Recommended Solution
```go
package pipeline

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"
)

type Item struct {
	ID   int
	Data string
}

func RunPipeline(ctx context.Context, totalItems int) error {
	// errgroup creates a context that cancels as soon as any stage returns an error
	g, ctx := errgroup.WithContext(ctx)

	readCh := make(chan Item)
	transformCh := make(chan Item)

	// Stage 1: Reader / Producer
	g.Go(func() error {
		defer close(readCh)
		for i := 0; i < totalItems; i++ {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case readCh <- Item{ID: i, Data: fmt.Sprintf("raw-%d", i)}:
			}
		}
		return nil
	})

	// Stage 2: Transformer
	g.Go(func() error {
		defer close(transformCh)
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case item, ok := <-readCh:
				if !ok {
					return nil
				}
				if item.ID == 13 {
					return fmt.Errorf("bad item id %d", item.ID)
				}
				select {
				case <-ctx.Done():
					return ctx.Err()
				case transformCh <- Item{ID: item.ID, Data: item.Data + "-transformed"}:
				}
			}
		}
	})

	// Stage 3: Uploader / Consumer
	g.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case item, ok := <-transformCh:
				if !ok {
					return nil
				}
				if err := uploadItem(ctx, item); err != nil {
					return fmt.Errorf("upload failed for item %d: %w", item.ID, err)
				}
			}
		}
	})

	// g.Wait() blocks until all stages finish and returns the first non-nil error
	if err := g.Wait(); err != nil {
		return fmt.Errorf("pipeline error: %w", err)
	}

	return nil
}

func uploadItem(ctx context.Context, item Item) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}
```


---

## MR #5: Distributed Rate Limiter & Token Bucket

### 1. MR Description
> *"This MR implements a Token Bucket rate limiter designed to limit concurrent API calls to an external third-party service (e.g., GitLab REST API / GitHub API). Callers call `Wait()` to block until a token is available or until context expires. We use a ticker goroutine to replenish tokens into a channel."*

### 2. The Code (Under Review)
```go
package ratelimit

import (
	"context"
	"errors"
	"sync"
	"time"
)

var ErrLimitExceeded = errors.New("rate limit exceeded or canceled")

type TokenBucket struct {
	tokens   chan struct{}
	capacity int
	rate     time.Duration // interval between tokens
	mu       sync.Mutex
	stopCh   chan struct{}
}

func NewTokenBucket(capacity int, refillInterval time.Duration) *TokenBucket {
	tb := &TokenBucket{
		tokens:   make(chan struct{}, capacity),
		capacity: capacity,
		rate:     refillInterval,
		stopCh:   make(chan struct{}),
	}

	// Fill bucket initially
	for i := 0; i < capacity; i++ {
		tb.tokens <- struct{}{}
	}

	// Refill goroutine
	go tb.startRefill()

	return tb
}

func (tb *TokenBucket) startRefill() {
	ticker := time.NewTicker(tb.rate)
	defer ticker.Stop()

	for {
		select {
		case <-tb.stopCh:
			return
		case <-ticker.C:
			// Add token to bucket if not full
			select {
			case tb.tokens <- struct{}{}:
			default:
				// Bucket is full, drop token
			}
		}
	}
}

// Wait blocks until a token is available or context is cancelled.
func (tb *TokenBucket) Wait(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-tb.tokens:
		return nil
	case <-tb.stopCh:
		return ErrLimitExceeded
	}
}

// Close stops the refill worker and cleans up resources.
func (tb *TokenBucket) Close() {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	close(tb.stopCh)
	close(tb.tokens)
}
```

### 3. Why It Is Bad (Flaws & Failure Modes)
1. **Closing `tokens` Causes `Wait()` to Bypass Rate Limits (The Fatal Trap):**
   - In Go, reading from a closed channel does not block; it immediately yields the zero value (`struct{}{}`).
   - In `Wait(ctx)`, both `case <-tb.tokens:` and `case <-tb.stopCh:` become ready upon `Close()`. Go selects randomly. If `case <-tb.tokens:` is picked, it returns `nil` (SUCCESS).
   - After shutdown, callers of `Wait()` bypass rate limits entirely at infinite speed instead of being rejected.
2. **Panic on Sending to Closed Channel in `startRefill()`:**
   - When `Close()` runs `close(tb.tokens)`, a concurrent tick in `startRefill()` executing `tb.tokens <- struct{}{}` triggers an immediate panic: `send on closed channel`.
3. **Double Close Panic & Useless Mutex:**
   - Calling `Close()` twice panics on `close(tb.stopCh)`. `tb.mu` does not track closure state. Use `sync.Once`.
4. **Ticker Panics on Non-Positive Durations:**
   - If `capacity <= 0` or `refillInterval <= 0`, `time.NewTicker` panics at runtime. Input validation is missing.
5. **Architectural Overhead (Active Ticker vs Lazy Calculation):**
   - Running a background goroutine and OS ticker per limiter does not scale when managing thousands of rate limiters (e.g. per-tenant/per-repo limiters). High rates (e.g. 10k RPS) exceed ticker granularity.

### 4. Ideal Review Comments to Leave on the MR
- **On `Close()` and `Wait()` Semantics:**
  > *"Closing `tb.tokens` causes `case <-tb.tokens:` in `Wait()` to immediately yield zero-value structs. Callers calling `Wait()` after `Close()` will receive `nil` (success) at infinite speed instead of an error, completely bypassing the rate limit. Do not close `tb.tokens`; signal termination solely via `stopCh`."*
- **On Panic in `startRefill()`:**
  > *"Closing `tb.tokens` creates a race condition with `startRefill()`. If `ticker.C` fires concurrently with `Close()`, `tb.tokens <- struct{}{}` will panic with `send on closed channel`."*
- **On `Close()` Idempotency & `tb.mu`:**
  > *"Calling `Close()` twice will panic on closing already-closed channels. `tb.mu` does not guard against this. Use `sync.Once` to ensure `Close()` is safe to call concurrently and repeatedly."*
- **On Scale / Architecture Recommendation:**
  > *"Active ticker goroutines introduce scheduler and timer overhead for large numbers of limiters. In production, consider a lazy token calculation based on elapsed time (`time.Now().Sub(lastRefill)`) as implemented in `golang.org/x/time/rate`."*

### 5. Recommended Solution
```go
package ratelimit

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrClosed        = errors.New("rate limiter is closed")
	ErrInvalidParams = errors.New("capacity and refillInterval must be greater than zero")
)

type TokenBucket struct {
	tokens chan struct{}
	stopCh chan struct{}
	once   sync.Once
}

func NewTokenBucket(capacity int, refillInterval time.Duration) (*TokenBucket, error) {
	if capacity <= 0 || refillInterval <= 0 {
		return nil, ErrInvalidParams
	}

	tb := &TokenBucket{
		tokens: make(chan struct{}, capacity),
		stopCh: make(chan struct{}),
	}

	// Fill bucket initially
	for i := 0; i < capacity; i++ {
		tb.tokens <- struct{}{}
	}

	// Refill goroutine
	go tb.startRefill(refillInterval)

	return tb, nil
}

func (tb *TokenBucket) startRefill(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-tb.stopCh:
			return
		case <-ticker.C:
			select {
			case tb.tokens <- struct{}{}:
			default:
				// Bucket is full, drop token
			}
		}
	}
}

// Wait blocks until a token is acquired, context is canceled, or limiter is closed.
func (tb *TokenBucket) Wait(ctx context.Context) error {
	select {
	case <-tb.stopCh:
		return ErrClosed
	case <-ctx.Done():
		return ctx.Err()
	case <-tb.tokens:
		return nil
	}
}

// Close gracefully stops the refill goroutine. It is safe to call multiple times.
func (tb *TokenBucket) Close() {
	tb.once.Do(func() {
		close(tb.stopCh)
	})
}
```


---

## MR #6: Event Listener & Notifier Registry

### 1. MR Description
> *"We implement an Event Notifier system where different service components can register event listeners. When an event occurs (e.g. `UserCreated`), the notifier dispatches the event concurrently to all registered listeners. We also support unregistering listeners."*

### 2. The Code (Under Review)
```go
package notifier

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Event struct {
	Type      string
	Payload   string
	Timestamp time.Time
}

type Listener interface {
	OnEvent(ctx context.Context, event *Event) error
}

type Notifier struct {
	mu        sync.Mutex
	listeners []Listener
}

func NewNotifier() *Notifier {
	return &Notifier{
		listeners: make([]Listener, 0),
	}
}

func (n *Notifier) Register(l Listener) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.listeners = append(n.listeners, l)
}

func (n *Notifier) Unregister(l Listener) {
	n.mu.Lock()
	defer n.mu.Unlock()

	for i, listener := range n.listeners {
		if listener == l {
			n.listeners = append(n.listeners[:i], n.listeners[i+1:]...)
			break
		}
	}
}

// Dispatch sends the event to all listeners concurrently with a timeout.
func (n *Notifier) Dispatch(event Event) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	n.mu.Lock()
	listeners := n.listeners
	n.mu.Unlock()

	var wg sync.WaitGroup

	for _, l := range listeners {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := l.OnEvent(ctx, &event)
			if err != nil {
				fmt.Printf("listener error: %v
", err)
			}
		}()
	}

	wg.Wait()
}
```

### 3. Why It Is Bad (Flaws & Failure Modes)
1. **Loop Variable Capture (`l`) in Goroutine Closure:**
   - In Go (especially pre-1.22 semantics), `l` is allocated once and reused across loop iterations.
   - When goroutines start, `l` frequently evaluates to the last element in the slice. The same listener gets called $N$ times while earlier listeners are never called.
2. **Shallow Slice Copy Race (`listeners := n.listeners`):**
   - In Go, assigning a slice copies only the 24-byte header (pointer, len, cap). The underlying array is shared.
   - If `Unregister()` executes concurrently, it mutates the underlying array via `append()`. Reading and modifying the same slice array concurrently without locks is a data race.
3. **Cross-Goroutine Data Race via Pointer (`&event`):**
   - Passing a shared pointer `&event` across multiple concurrent goroutines means any listener modifying the event causes a data race or unpredictable state for other listeners.
4. **Memory Leak in `Unregister`:**
   - Deleting an element from a slice of interfaces or pointers using `append(s[:i], s[i+1:]...)` leaves the last element in the underlying array alive. The garbage collector cannot free the unregistered listener object.
5. **Context Decoupling (`context.Background()`):**
   - Hardcoding `context.Background()` severs caller context cancellation, deadlines, and distributed tracing spans.

### 4. Ideal Review Comments to Leave on the MR
- **On Closure Variable Capture:**
  > *"The loop variable `l` is captured by reference in the goroutine closure. Pass `l` explicitly into the anonymous function (`go func(target Listener) { ... }(l)`) to prevent all workers invoking the same captured instance."*
- **On Shallow Slice Copy Race:**
  > *"Doing `listeners := n.listeners` copies the slice header, but both slices still reference the exact same backing array. If `Unregister()` modifies the slice while `Dispatch()` is reading, this triggers a data race. Create a separate snapshot using `make([]Listener, len(n.listeners))` and `copy()` under lock."*
- **On Pointer Passing `&event`:**
  > *"Passing the same pointer `&event` to multiple concurrent listeners allows one listener to mutate the event data while others are reading it, causing a data race. Pass events by value or create per-goroutine copies."*
- **On Memory Leak in `Unregister`:**
  > *"When deleting from `n.listeners`, the shifted final element still holds an interface reference in the underlying backing array, preventing garbage collection. Set `n.listeners[last] = nil` before re-slicing."*
- **On Context Propagation:**
  > *"Avoid hardcoding `context.Background()` in internal methods. Accept `ctx context.Context` from the caller so that request cancellations and distributed tracing span contexts propagate properly."*

### 5. Recommended Solution
```go
package notifier

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Event struct {
	Type      string
	Payload   string
	Timestamp time.Time
}

type Listener interface {
	OnEvent(ctx context.Context, event Event) error
}

type Notifier struct {
	mu        sync.RWMutex
	listeners []Listener
}

func NewNotifier() *Notifier {
	return &Notifier{
		listeners: make([]Listener, 0),
	}
}

func (n *Notifier) Register(l Listener) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.listeners = append(n.listeners, l)
}

func (n *Notifier) Unregister(l Listener) {
	n.mu.Lock()
	defer n.mu.Unlock()

	for i, listener := range n.listeners {
		if listener == l {
			last := len(n.listeners) - 1
			copy(n.listeners[i:], n.listeners[i+1:])
			n.listeners[last] = nil // Clear reference so GC can reclaim memory
			n.listeners = n.listeners[:last]
			break
		}
	}
}

// Dispatch sends events to all listeners concurrently, deriving timeout from caller context.
func (n *Notifier) Dispatch(ctx context.Context, event Event) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// Snapshot listeners to avoid holding lock and prevent data race on slice backing array
	n.mu.RLock()
	listeners := make([]Listener, len(n.listeners))
	copy(listeners, n.listeners)
	n.mu.RUnlock()

	var wg sync.WaitGroup

	for _, l := range listeners {
		wg.Add(1)
		// Explicitly pass listener as argument to avoid closure variable capture issues
		go func(target Listener) {
			defer wg.Done()
			// Pass event by value to prevent cross-goroutine mutation races
			if err := target.OnEvent(ctx, event); err != nil {
				fmt.Printf("listener error: %v
", err)
			}
		}(l)
	}

	wg.Wait()
}
```
