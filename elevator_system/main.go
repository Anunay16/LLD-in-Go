package main

import (
	"elevator_system/elevator"
	"fmt"
)

func main() {
	elevator1 := elevator.NewElevator(0, elevator.DefaultConfig())
	fmt.Println(elevator1.GetState())
}
