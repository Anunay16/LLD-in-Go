package main

import (
	"fmt"
	"sync"
)

func main() {
	const rounds = 3
	chanA := make(chan struct{})
	chanB := make(chan struct{})
	chanC := make(chan struct{})

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		for i := 0; i < rounds; i++ {
			<-chanA
			fmt.Print("A")
			chanB <- struct{}{}
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < rounds; i++ {
			<-chanB
			fmt.Print("B")
			chanC <- struct{}{}
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < rounds; i++ {
			<-chanC
			fmt.Println("C")
			// Only signal A if there is a next round
			if i < rounds-1 {
				chanA <- struct{}{}
			}
		}
	}()

	// Start the cycle
	chanA <- struct{}{}
	wg.Wait()
}
