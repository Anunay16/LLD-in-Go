package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// TaskTracker monitors concurrent task executions using atomic operations to verify bounds.
type TaskTracker struct {
	activeCount int32
	maxObserved int32
}

func (t *TaskTracker) StartTask(taskID int) {
	current := atomic.AddInt32(&t.activeCount, 1)
	for {
		max := atomic.LoadInt32(&t.maxObserved)
		if current <= max || atomic.CompareAndSwapInt32(&t.maxObserved, max, current) {
			break
		}
	}
	fmt.Printf("[Task %2d] STARTED  (Currently active: %d)\n", taskID, current)
}

func (t *TaskTracker) EndTask(taskID int) {
	current := atomic.AddInt32(&t.activeCount, -1)
	fmt.Printf("[Task %2d] FINISHED (Currently active: %d)\n", taskID, current)
}

func simulateTask(taskID int, tracker *TaskTracker) {
	tracker.StartTask(taskID)
	// Simulate work / network I/O
	time.Sleep(100 * time.Millisecond)
	tracker.EndTask(taskID)
}

// RunBoundedTasks executes totalTasks while ensuring no more than maxConcurrent
// tasks run simultaneously.
func RunBoundedTasks(totalTasks int, maxConcurrent int) {
	tracker := &TaskTracker{}
	var wg sync.WaitGroup

	// TODO 1: Create a semaphore channel with capacity equal to maxConcurrent.
	// Hint: A buffered channel of struct{} acts as a counting semaphore.
	sem := make(chan struct{}, maxConcurrent)

	for i := 1; i <= totalTasks; i++ {
		taskID := i
		wg.Add(1)

		// TODO 2: Acquire a semaphore slot before (or inside) running the goroutine.
		// Think carefully: should you acquire BEFORE launching the goroutine or INSIDE it?
		// What is the trade-off in terms of goroutine allocation?
		sem <- struct {}{}

		go func() {
			defer func() {
				<- sem
				wg.Done()
			}()

			// TODO 3: Acquire slot here (if not acquired outside)

			simulateTask(taskID, tracker)

			// TODO 4: Release the slot when the task finishes (hint: defer)
		}()
	}

	wg.Wait()

	// --- Automated Verification Check ---
	fmt.Println("\n--- Execution Report ---")
	fmt.Printf("Concurrency limit:             %d\n", maxConcurrent)
	fmt.Printf("Max concurrent tasks observed: %d\n", tracker.maxObserved)
	if tracker.maxObserved == 0 {
		fmt.Println("Result: ❌ No tasks were executed.")
	} else if tracker.maxObserved <= int32(maxConcurrent) {
		fmt.Println("Result: ✅ SUCCESS! Concurrency stayed within the bounds.")
	} else {
		fmt.Println("Result: ❌ FAILED! Concurrency exceeded the limit.")
	}
}

func main() {
	totalTasks := 10
	maxConcurrent := 3

	fmt.Printf("Running %d tasks with at most %d concurrent tasks allowed...\n\n", totalTasks, maxConcurrent)
	RunBoundedTasks(totalTasks, maxConcurrent)
}
