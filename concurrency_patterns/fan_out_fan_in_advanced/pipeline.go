package main

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Task represents a URL health-check job.
type Task struct {
	ID  int
	URL string
}

// Result represents the outcome of a health-check.
type Result struct {
	TaskID     int
	WorkerID   int
	StatusCode int
	Err        error
}

// generateTasks produces tasks onto a channel.
func generateTasks(ctx context.Context, count int) <-chan Task {
	out := make(chan Task)
	go func() {
		defer close(out)
		for i := 1; i <= count; i++ {
			url := fmt.Sprintf("https://api.service.internal/item/%d", i)
			// Simulate failing URLs on specific IDs
			if i == 5 || i == 8 || i == 12 {
				url = fmt.Sprintf("https://api.service.internal/fail/%d", i)
			}
			select {
			case out <- Task{ID: i, URL: url}:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

// worker reads from the shared tasks channel, processes health checks, and produces results.
func worker(ctx context.Context, id int, tasks <-chan Task) <-chan Result {
	out := make(chan Result)

	go func() {
		defer close(out)

		// TODO 1: Continuously pull tasks from the shared `tasks` channel until closed or ctx.Done().
		// For each task:
		// 1. Simulate work: time.Sleep(40 * time.Millisecond)
		// 2. If task.URL contains "fail", simulate an error:
		//    res := Result{TaskID: task.ID, WorkerID: id, StatusCode: 500, Err: errors.New("HTTP 500 Internal Error")}
		//    Otherwise:
		//    res := Result{TaskID: task.ID, WorkerID: id, StatusCode: 200, Err: nil}
		// 3. Send `res` to `out`.
		//    CRITICAL: You MUST use select with `case out <- res:` and `case <-ctx.Done():`
		//    so that if downstream cancels, this worker never gets blocked on a full channel!

		for {
			select {
			case <-ctx.Done():
				return
			case task, ok := <-tasks:
				if !ok {
					return
				}
				time.Sleep(40 * time.Millisecond)
				var res Result
				if strings.Contains(task.URL, "fail") {
					res = Result{TaskID: task.ID, WorkerID: id, StatusCode: 500, Err: errors.New("HTTP 500 Internal Error")}
				} else {
					res = Result{TaskID: task.ID, WorkerID: id, StatusCode: 200, Err: nil}
				}

				// TODO: Send res to out, but respect ctx.Done()!
				select {
				case <-ctx.Done():
					return
				case out <- res:
				}
			}
		}
	}()

	return out
}

// mergeWithContext multiplexes multiple Result channels into a single channel with context cancellation.
func mergeWithContext(ctx context.Context, channels ...<-chan Result) <-chan Result {
	out := make(chan Result)
	var wg sync.WaitGroup

	// TODO 2: For each channel in `channels`:
	// 1. Spawn a forwarding goroutine with wg.Add(1) and defer wg.Done().
	// 2. Loop through the channel. On each result, send to `out`.
	//    CRITICAL: Use select to listen to BOTH `out <- res` AND `<-ctx.Done()`.
	//    If ctx.Done() is received, exit the forwarding goroutine immediately!
	for _, ch := range channels {
		wg.Add(1)
		go func(c <-chan Result) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case res, ok := <-c:
					if !ok {
						return
					}
					select {
					case out <- res:
					case <-ctx.Done():
						return
					}
				}
			}
		}(ch)
	}

	// TODO 3: In a separate background goroutine, wait for wg.Wait() or ctx.Done(),
	// and ensure `close(out)` is called safely so the consumer stops reading.
	go func() {
		wg.Wait()
		close(out)
	}()

	// Temporary: closes out immediately so the starter template compiles and runs
	// close(out)

	return out
}

func main() {
	const totalTasks = 20
	const numWorkers = 4
	const maxErrorsBeforeAbort = 2 // Pipeline will cancel early if 2 errors occur!

	baselineGoroutines := runtime.NumGoroutine()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fmt.Printf("Starting Advanced Pipeline: %d tasks, %d workers, abort on %d errors...\n\n",
		totalTasks, numWorkers, maxErrorsBeforeAbort)

	// 1. Generate task stream
	taskStream := generateTasks(ctx, totalTasks)

	// 2. Fan-Out: Spawn 4 workers sharing the same taskStream
	workerChans := make([]<-chan Result, numWorkers)
	for i := 0; i < numWorkers; i++ {
		workerChans[i] = worker(ctx, i+1, taskStream)
	}

	// 3. Fan-In: Merge all worker channels into a single output channel
	mergedResults := mergeWithContext(ctx, workerChans...)

	// 4. Consumer: Read results, and cancel context if error threshold is reached
	successCount := 0
	errorCount := 0

	for res := range mergedResults {
		if res.Err != nil {
			errorCount++
			fmt.Printf("⚠️  [Worker %d] Task %2d FAILED: %v (Errors so far: %d)\n",
				res.WorkerID, res.TaskID, res.Err, errorCount)
			if errorCount >= maxErrorsBeforeAbort {
				fmt.Println("\n🛑 Error threshold reached! Cancelling context to abort all workers...")
				cancel()
				break
			}
		} else {
			successCount++
			fmt.Printf("✅ [Worker %d] Task %2d SUCCESS: HTTP %d\n",
				res.WorkerID, res.TaskID, res.StatusCode)
		}
	}

	// Allow goroutines a moment to clean up and exit
	time.Sleep(150 * time.Millisecond)

	// --- Automated Goroutine Leak Check ---
	activeGoroutines := runtime.NumGoroutine()
	leaked := activeGoroutines - baselineGoroutines

	fmt.Println("\n--- Pipeline Verification Report ---")
	fmt.Printf("Successful tasks processed: %d\n", successCount)
	fmt.Printf("Failed tasks encountered:   %d\n", errorCount)
	fmt.Printf("Baseline Goroutines:        %d\n", baselineGoroutines)
	fmt.Printf("Active Goroutines now:      %d\n", activeGoroutines)

	if leaked > 1 {
		fmt.Printf("Result: ❌ FAILED! Leaked %d goroutines! Upstream workers are hanging on blocked sends.\n", leaked)
	} else if errorCount >= maxErrorsBeforeAbort {
		fmt.Println("Result: ✅ SUCCESS! Early cancellation aborted pipeline cleanly with 0 leaked goroutines.")
	} else {
		fmt.Println("Result: ❌ FAILED! Expected pipeline to encounter errors and test early abort.")
	}
}
