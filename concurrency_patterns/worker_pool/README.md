# Problem 2: Worker Pool with Graceful Shutdown

## Problem Statement
Implement a concurrent **Worker Pool** in Go. 

You have a queue of `numJobs` tasks that need to be processed by a fixed pool of `numWorkers` goroutines (where `numWorkers < numJobs`). Each worker should pull jobs concurrently, process them, and send back results to a results channel.

```
       +----------+      +-----------+
       | Producer | ---> | Job Queue |
       +----------+      +-----------+
                               |
               +---------------+---------------+
               |               |               |
         +----------+    +----------+    +----------+
         | Worker 1 |    | Worker 2 |    | Worker 3 |
         +----------+    +----------+    +----------+
               |               |               |
               +---------------+---------------+
                               |
                               v
                       +--------------+
                       | Result Queue |
                       +--------------+
                               |
                               v
                       +--------------+
                       |  Collector   |
                       +--------------+
```

---

## Requirements
1. **Fixed Concurrency**: Exactly `numWorkers` worker goroutines must be created to process all jobs.
2. **Channel Ownership & Closure**:
   - The producer must send all jobs and close the `jobs` channel when done.
   - When all workers finish processing (using `sync.WaitGroup`), the `results` channel must be closed safely.
3. **No Deadlocks or Leaks**:
   - Ensure workers do not block indefinitely after all jobs are processed.
   - Ensure the results collector doesn't block waiting for more results once all workers exit.
4. **Result Verification**:
   - All `numJobs` must be accounted for in the output, showing which worker processed each job.

---

## Key Go Rules to Keep in Mind
- **Rule of thumb for closing channels**: The sender should close the channel, never the receiver.
- A `for job := range jobs` loop automatically exits when `jobs` is closed and empty.
- To close `results`, wait for all workers to finish in a separate goroutine or after worker `wg.Wait()`.

---

## Running the Code
```bash
go run worker_pool/worker_pool.go
```

