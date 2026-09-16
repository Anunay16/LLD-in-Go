# Advanced Problem: Fan-Out, Fan-In Pipeline with Context Cancellation & Error Propagation

## Problem Statement
In production systems, data pipelines are rarely as simple as fixed numbers passing through happy-path workers. You must handle:
1. **Dynamic Fan-Out**: Sharing a single input stream among $W$ concurrent workers.
2. **Failures and Errors**: Workers can fail (e.g. HTTP 500 or timeout), returning `(Data, Err)` structs.
3. **Context Cancellation & Early Teardown**: Downstream consumers might stop early (e.g. after finding a solution or hitting an error threshold). When that happens, all upstream workers must stop immediately without hanging or leaking goroutines.

```
       +------------------+
       |  Task Generator  | (Streams 20 tasks, some valid, some failing)
       +------------------+
                 |
                 v  (tasks chan)
       +---------+---------+---------+
       |         |         |         |
       v         v         v         v
   [Worker 1] [Worker 2] [Worker 3] [Worker 4]  <-- Fan-Out (W workers)
       |         |         |         |
      ch1       ch2       ch3       ch4
       |         |         |         |
       +---------+---------+---------+
                 |
                 v
   +---------------------------+
   |    Merge with Context     |  <-- Fan-In (Multiplexer with ctx.Done())
   +---------------------------+
                 |
                 v  (merged results chan)
         +---------------+
         |   Consumer    | (Cancels context if 2 errors occur or done)
         +---------------+
```

---

## Requirements

### 1. Worker Function (`worker`)
Each worker should read tasks from the shared `tasks` channel:
- Must respect `ctx.Done()`: if context is canceled, exit immediately!
- When sending results to its dedicated output channel, must use `select`:
  ```go
  select {
  case out <- result:
  case <-ctx.Done():
      return
  }
  ```
- Close its output channel when done or upon cancellation.

### 2. Fan-In Multiplexer (`mergeWithContext`)
```go
func mergeWithContext(ctx context.Context, channels ...<-chan Result) <-chan Result
```
- Must read from all input channels concurrently and forward to `out`.
- Must listen for `ctx.Done()`. If canceled, terminate forwarding immediately without blocking.
- Must close `out` only when all input channels are drained, or immediately if `ctx` is canceled.

### 3. Goroutine Leak Prevention
- A common trap in Go pipelines: if the consumer cancels `ctx` and exits, workers remain blocked trying to write to unread channels!
- The automated verification harness at the end checks `runtime.NumGoroutine()` before and after to verify that **zero goroutines leaked**.

---

## Running the Code
```bash
go run -race fan_out_fan_in_advanced/pipeline.go
```

