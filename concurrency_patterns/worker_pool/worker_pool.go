package main

import (
	"fmt"
	"sync"
	"time"
)

// Job represents a unit of work.
type Job struct {
	ID   int
	Data int
}

// Result represents the outcome of processing a Job.
type Result struct {
	Job       Job
	WorkerID  int
	Output    int
	Processed time.Time
}

// worker processes jobs from the jobs channel and sends results to the results channel.
func worker(id int, jobs <-chan Job, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()

	// TODO: Continuously read jobs from the jobs channel until it is closed.
	// For each job:
	// 1. Simulate some processing time (e.g. time.Sleep(50 * time.Millisecond)).
	// 2. Compute the result (e.g. output = job.Data * 2).
	// 3. Send a Result instance to the results channel.
	for job := range jobs {
		time.Sleep(50 * time.Millisecond)
		job.Data = job.Data * 2
		results <- Result{Job: job, WorkerID: id, Output: job.Data, Processed: time.Now()}
	}
}

func main() {
	const numJobs = 15
	const numWorkers = 3
	const bufferSize = 3 // Constant small buffer: memory usage stays O(1) regardless of numJobs!

	fmt.Printf("Starting Streaming Worker Pool (%d workers, %d jobs, buffer size %d)...\n\n",
		numWorkers, numJobs, bufferSize)

	jobs := make(chan Job, bufferSize)
	results := make(chan Result, bufferSize)

	var wg sync.WaitGroup

	// 1. Start Worker goroutines
	for i := 1; i <= numWorkers; i++ {
		wg.Add(1)
		go worker(i, jobs, results, &wg)
	}

	// 2. PRODUCER: Enqueue jobs in a background goroutine and close jobs when done
	go func() {
		for i := 1; i <= numJobs; i++ {
			jobs <- Job{ID: i, Data: i}
		}
		close(jobs) // Signals workers that no more jobs will be produced
	}()

	// 3. CLOSER: Wait for workers to finish in background, then close results
	go func() {
		wg.Wait()
		close(results) // Signals consumer that all processing is done
	}()

	// 4. CONSUMER: Stream and process results in real-time in main
	for res := range results {
		fmt.Printf("[Worker %d] Processed Job %2d -> Output: %2d (at %s)\n",
			res.WorkerID, res.Job.ID, res.Output, res.Processed.Format("15:04:05.000"))
	}

	fmt.Println("\nAll jobs successfully streamed and consumed!")
}
