package main

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

// generateNumbers sends a list of integers onto a read-only channel and closes it when done.
func generateNumbers(nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, n := range nums {
			out <- n
		}
	}()
	return out
}

// squareWorker takes an input channel of numbers and squares each number concurrently.
func squareWorker(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			// Simulate processing time
			time.Sleep(20 * time.Millisecond)
			out <- n * n
		}
	}()
	return out
}

// merge multiplexes multiple input channels into a single output channel (Fan-In).
func merge(channels ...<-chan int) <-chan int {
	out := make(chan int)

	var wg sync.WaitGroup

	// TODO 1: For each channel in `channels`, start a goroutine that reads every
	// value from that channel and forwards it into `out`.
	// Use wg.Add(1) and defer wg.Done() for each goroutine.
	for _, ch := range channels {
		wg.Add(1)
		go func(ch <-chan int) {
			defer wg.Done()
			for val := range ch {
				out <- val
			}
		}(ch)
	}

	// TODO 2: In a separate goroutine, wait for all forwarding goroutines to finish (wg.Wait()),
	// and then close `out`. This signals to the consumer that all streams are done.
	go func ()  {
		wg.Wait()
		close(out)
	} ()

	// NOTE: Remove the line below once you start implementing the TODOs:
	// close(out)

	return out
}

func main() {
	// 1. Generate two streams of numbers
	stream1 := generateNumbers(1, 2, 3)
	stream2 := generateNumbers(4, 5, 6)

	// 2. Fan-Out: Process each stream with independent square workers
	worker1 := squareWorker(stream1)
	worker2 := squareWorker(stream2)

	// 3. Fan-In: Merge both worker output streams into a single channel
	fmt.Println("Multiplexing streams with merge()...")
	merged := merge(worker1, worker2)

	// 4. Collect results from the merged stream
	var results []int
	for val := range merged {
		results = append(results, val)
	}

	// 5. Verification
	expected := []int{1, 4, 9, 16, 25, 36}
	sort.Ints(results)

	fmt.Println("\n--- Verification Report ---")
	fmt.Printf("Expected results: %v\n", expected)
	fmt.Printf("Received results: %v\n", results)

	if fmt.Sprint(results) == fmt.Sprint(expected) {
		fmt.Println("Result: ✅ SUCCESS! All items merged and collected correctly.")
	} else {
		fmt.Println("Result: ❌ FAILED! Merged stream did not produce the expected items.")
	}
}
