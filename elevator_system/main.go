package main

import (
	"fmt"

	"elevator_system/elevator"
)

func main() {
	// Initialize ElevatorSystem with 3 elevators serving floors 0-9
	sys := elevator.NewElevatorSystem(nil, nil)
	fmt.Println("=== Elevator Simulation Started (3 Elevators, Floors 0-9) ===")

	// User on floor 3 makes a hall call going UP
	fmt.Println("\n[Action] Hall call from floor 3 (Direction: UP)")
	sys.RequestElevator(3, elevator.Up)

	// Step simulation until hall call is served
	for step := 1; step <= 3; step++ {
		sys.Step()
		printStatus(sys, step)
	}

	// Passenger enters Elevator 0 and selects destination floors 5 and 7
	fmt.Println("\n[Action] Passenger inside Elevator 0 selects floors 5 and 7")
	sys.SelectDestination(0, 5)
	sys.SelectDestination(0, 7)

	// Step simulation until destination floors are served
	for step := 4; step <= 8; step++ {
		sys.Step()
		printStatus(sys, step)
	}

	// Demonstrate invalid request rejection (Req 7)
	fmt.Println("\n[Validation Test] Requesting non-existent floor 10:")
	fmt.Printf("sys.RequestElevator(10, Down) = %v\n", sys.RequestElevator(10, elevator.Down))

	// Demonstrate current floor no-op (Req 8)
	fmt.Printf("sys.SelectDestination(0, 7) [already at 7] = %v\n", sys.SelectDestination(0, 7))
}

func printStatus(sys *elevator.ElevatorSystem, step int) {
	fmt.Printf("--- Step %d ---\n", step)
	for _, e := range sys.GetElevators() {
		fmt.Printf("Elevator %d | Floor: %d | State: %-10s | Direction: %-4s\n",
			e.GetID(), e.GetCurrentFloor(), e.GetState(), e.GetDirection())
	}
}
