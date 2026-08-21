package main

import "fmt"

import "state_pattern/vending_machine"

func main() {
	fmt.Println("==========================================")
	fmt.Println("     State Pattern Demo: Vending Machine  ")
	fmt.Println("==========================================")

	// Initialize Vending Machine with 2 items
	vm := vending_machine.NewVendingMachine(2)
	vm.DisplayCurrentState()

	fmt.Println("\n[Scenario 1] Normal Purchase:")
	if err := vm.InsertCoin(); err != nil {
		fmt.Printf("ERROR: %v\n", err)
	}
	if err := vm.SelectProduct(); err != nil {
		fmt.Printf("ERROR: %v\n", err)
	}
	vm.DisplayCurrentState()

	fmt.Println("\n[Scenario 2] Insert Coin and Eject:")
	if err := vm.InsertCoin(); err != nil {
		fmt.Printf("ERROR: %v\n", err)
	}
	if err := vm.EjectCoin(); err != nil {
		fmt.Printf("ERROR: %v\n", err)
	}
	vm.DisplayCurrentState()

	fmt.Println("\n[Scenario 3] Invalid Action (Selecting product without coin):")
	if err := vm.SelectProduct(); err != nil {
		fmt.Printf("ERROR: %v\n", err)
	}

	fmt.Println("\n[Scenario 4] Purchase Final Item:")
	if err := vm.InsertCoin(); err != nil {
		fmt.Printf("ERROR: %v\n", err)
	}
	if err := vm.SelectProduct(); err != nil {
		fmt.Printf("ERROR: %v\n", err)
	}
	vm.DisplayCurrentState()

	fmt.Println("\n[Scenario 5] Action on Sold-out Machine:")
	if err := vm.InsertCoin(); err != nil {
		fmt.Printf("ERROR: %v\n", err)
	}
	if err := vm.SelectProduct(); err != nil {
		fmt.Printf("ERROR: %v\n", err)
	}
}
