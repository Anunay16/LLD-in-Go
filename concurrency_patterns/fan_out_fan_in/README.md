# Problem 4: Fan-Out, Fan-In Stream Multiplexing

## Problem Statement
In concurrent data pipelines, work is often broken down into stages:
- **Fan-Out**: Multiple workers read from a data source to perform intensive computations in parallel. Each worker produces results onto its own channel.
- **Fan-In**: A multiplexer merges results from multiple independent channels into a **single consolidated output channel**.

```
                         +------------+
                   +---> | Producer 1 | --- ch1 ---+
                   |     +------------+            |
                   |                               v
             +----------+                 +-----------------+      +----------+
  Input ---> | Fan-Out  |                 | Fan-In (Merge)  | ---> | Consumer |
             +----------+                 +-----------------+      +----------+
                   |                               ^
                   |     +------------+            |
                   +---> | Producer 2 | --- ch2 ---+
                         +------------+
```

---

## Your Goal
Implement the `merge` (Fan-In) function:
```go
func merge(channels ...<-chan int) <-chan int
```

Given any number of input channels (`channels ...<-chan int`), `merge` must:
1. Return a single output channel `<-chan int`.
2. Read all values from each input channel concurrently and forward them into the output channel.
3. **Graceful Closing**: Once all input channels have been completely drained and closed, the output channel must be closed automatically.
4. Prevent goroutine leaks: No worker should be left hanging or blocked on an unread channel.

---

## Running the Code
```bash
go run fan_out_fan_in/fan_in.go
```

