# Problem 3: Bounded Concurrency / Semaphore Pattern

## Problem Statement
Suppose you have `N` asynchronous tasks (e.g. scraping URLs, calling third-party APIs, or querying databases). 
If you launch all `N` tasks simultaneously using `go doTask()`, you risk overwhelming the system, getting rate-limited (HTTP 429), or exhausting file descriptors.

Your goal is to execute all `N` tasks asynchronously while guaranteeing that **at most `K` tasks run concurrently at any given instant** (e.g., `N = 10`, `K = 3`).

---

## What is a Semaphore in Go?
A counting semaphore can be implemented cleanly in Go using a **buffered channel**:
- A buffered channel of capacity `K`: `sem := make(chan struct{}, K)`
- **Acquire token** (before starting work): `sem <- struct{}{}`
  - If `K` tokens are already in the channel, the `send` blocks until another goroutine reads from it!
- **Release token** (after finishing work): `<-sem`
  - Frees up a slot in the buffer for another waiting goroutine.

---

## Requirements
1. Process all `totalTasks` (e.g., 10 tasks).
2. Ensure no more than `maxConcurrent` (e.g., 3) tasks are actively executing at the same time.
3. Use a `sync.WaitGroup` to wait until all tasks have finished.
4. The built-in atomic verification tracker must report `Max concurrent tasks observed <= maxConcurrent`.

---

## Running the Code
```bash
go run bounded_runner/runner.go
```

