package main

import "fmt"

func main() {
	myChannel := make(chan string)

	go func() {
		myChannel <- "first"
		myChannel <- "second"
	}()

	msg1 := <-myChannel
	fmt.Printf("%s", msg1)
}
