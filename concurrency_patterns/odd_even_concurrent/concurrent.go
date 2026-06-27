package main

import (
	"fmt"
	"sync"
)

// FOR SELECT BASED APPROACH, MORE POWERFUL, BUT HERE
// USING ONLY ONE CASE INSIDE SELECT, SO DON'T MAKE SENSE

// func main() {
// 	arr := []int{1, 2, 3, 4, 5, 6}

// 	oddTurn := make(chan struct{})
// 	evenTurn := make(chan struct{})

// 	var wg sync.WaitGroup
// 	wg.Add(2)

// 	go func() {
// 		defer wg.Done()

// 		i := 0
// 		for {
// 			select {
// 			case _, ok := <-oddTurn:
// 				if !ok {
// 					return
// 				}

// 				if i >= len(arr) {
// 					close(evenTurn)
// 					return
// 				}

// 				fmt.Println(arr[i])
// 				i += 2

// 				evenTurn <- struct{}{}
// 			}
// 		}
// 	}()

// 	go func() {
// 		defer wg.Done()

// 		i := 1
// 		for {
// 			select {
// 			case _, ok := <-evenTurn:
// 				if !ok {
// 					return
// 				}

// 				if i >= len(arr) {
// 					close(oddTurn)
// 					return
// 				}

// 				fmt.Println(arr[i])
// 				i += 2

// 				oddTurn <- struct{}{}
// 			}
// 		}
// 	}()

// 	oddTurn <- struct{}{}
// 	wg.Wait()
// }

// FOR RNAGE LOOP BASED APPROACH: EASIER, RECOMMENDED FOR THIS PROBLEM

func main() {

	arr := []int{1, 2, 3, 4, 5, 6}
	oddTurn := make(chan struct{})
	evenTurn := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)
	// Prints 1,3,5...
	go func() {
		defer wg.Done()
		for i := 0; i < len(arr); i += 2 {
			<-oddTurn // Wait for my turn
			fmt.Println(arr[i])
			// Signal even goroutine if another even element exists
			if i+1 < len(arr) {
				evenTurn <- struct{}{}
			}
		}
	}()
	// Prints 2,4,6...
	go func() {
		defer wg.Done()
		for i := 1; i < len(arr); i += 2 {
			<-evenTurn // Wait for my turn
			fmt.Println(arr[i])
			// Signal odd goroutine if another odd element exists
			if i+2 < len(arr) {
				oddTurn <- struct{}{}
			}
		}
	}()
	// Start with odd goroutine
	oddTurn <- struct{}{}
	wg.Wait()

}
