package main

import (
	"fmt"
	"sync"
)

// RunConcurrentFizzBuzz coordinates 4 goroutines to print FizzBuzz from 1 to n in order.
func RunConcurrentFizzBuzz(n int) {
	// TODO: Create the channels or synchronization primitives needed to coordinate
	// the 4 goroutines (e.g. channels for fizz, buzz, fizzbuzz, and number).
	fizzChan := make(chan int)
	buzzChan := make(chan int)
	fizzBuzzChan := make(chan int)
	numChan := make(chan int)
	doneChan := make(chan struct{})

	var wg sync.WaitGroup
	wg.Add(4)

	// 1. Fizz Goroutine (prints "Fizz " for numbers divisible by 3 and not 5)
	go func() {
		defer wg.Done()
		// TODO: Listen for your turn, print "Fizz ", and coordinate with the next turn
		for range fizzChan {
			fmt.Print("Fizz ")
			doneChan <- struct{}{}
		}
	}()

	// 2. Buzz Goroutine (prints "Buzz " for numbers divisible by 5 and not 3)
	go func() {
		defer wg.Done()
		// TODO: Listen for your turn, print "Buzz ", and coordinate with the next turn
		for range buzzChan {
			fmt.Print("Buzz ")
			doneChan <- struct{}{}
		}
	}()

	// 3. FizzBuzz Goroutine (prints "FizzBuzz " for numbers divisible by both 3 and 5)
	go func() {
		defer wg.Done()
		// TODO: Listen for your turn, print "FizzBuzz ", and coordinate with the next turn
		for range fizzBuzzChan {
			fmt.Print("FizzBuzz ")
			doneChan <- struct{}{}
		}
	}()

	// 4. Number Goroutine (prints the number itself if not divisible by 3 or 5)
	go func() {
		defer wg.Done()
		// TODO: Listen for your turn, print the number, and coordinate with the next turn
		for i := range numChan {
			fmt.Printf("%d ", i)
			doneChan <- struct{}{}
		}
	}()

	// TODO: Drive the sequence from 1 to n (or start the initial token)
	// and ensure clean termination so all goroutines finish and no goroutines leak.

	for i:=1; i<=n; i++ {
		switch{
		case i%15==0:
			fizzBuzzChan <- i
		case i%3==0:
			fizzChan <- i
		case i%5==0:
			buzzChan <- i
		default:
			numChan <- i
		}
		<-doneChan
	}

	close(fizzChan)
	close(buzzChan)
	close(fizzBuzzChan)
	close(numChan)
	
	wg.Wait()
	fmt.Println()
}

func main() {
	n := 15
	fmt.Printf("Running Concurrent FizzBuzz up to %d:\n", n)
	fmt.Println("Expected: 1 2 Fizz 4 Buzz Fizz 7 8 Fizz Buzz 11 Fizz 13 14 FizzBuzz")
	fmt.Print("Actual:   ")
	RunConcurrentFizzBuzz(n)
}
