package main

import "fmt"

func checkDeferVarInstance() {
	x := 10
	defer fmt.Printf("Printing from defer: %d\n", x)
	x++
	fmt.Printf("Printing from main: %d\n", x)

	/*
	OP:
		Printing from main: 11
		Printing from defer: 10

	Which indicates that defer is storing the immidiate value of x
	*/
}

func checkMultipleDefers() {
	defer fmt.Println("1")
	defer fmt.Println("2")
	defer fmt.Println("3")
	/*	
	OP:
		3
		2
		1
	Which indicates that defers are executed in the reverse order
	*/
}

func main() {
	// checkDeferVarInstance()
	checkMultipleDefers()
}
