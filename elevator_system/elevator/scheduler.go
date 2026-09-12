package elevator

import "math"

// Scheduler defines the Strategy interface for assigning hall calls to elevators.
type Scheduler interface {
	SelectElevator(elevators []*Elevator, floor int, dir Direction) *Elevator
}

// NearestElevatorScheduler implements Scheduler using proximity and direction heuristics.
type NearestElevatorScheduler struct{}

func NewNearestElevatorScheduler() *NearestElevatorScheduler {
	return &NearestElevatorScheduler{}
}

// SelectElevator chooses the best elevator to serve a hall call.
func (s *NearestElevatorScheduler) SelectElevator(elevators []*Elevator, floor int, dir Direction) *Elevator {
	if len(elevators) == 0 {
		return nil
	}

	var bestElevator *Elevator
	bestCost := math.MaxInt32

	for _, e := range elevators {
		cost := s.calculateCost(e, floor, dir)
		if cost < bestCost {
			bestCost = cost
			bestElevator = e
		}
	}

	return bestElevator
}

func (s *NearestElevatorScheduler) calculateCost(e *Elevator, floor int, dir Direction) int {
	currentFloor := e.GetCurrentFloor()
	state := e.GetState()
	currentDir := e.GetDirection()

	// If already at the floor and idle, lowest possible cost
	if currentFloor == floor && state == StateIdle {
		return 0
	}

	switch currentDir {
	case Up:
		if floor >= currentFloor && dir == Up {
			// On the way up in the same direction
			return floor - currentFloor
		}
		// Must finish upward run, then reverse
		return (MaxFloor - currentFloor) + (MaxFloor - floor) + 2

	case Down:
		if floor <= currentFloor && dir == Down {
			// On the way down in the same direction
			return currentFloor - floor
		}
		// Must finish downward run, then reverse
		return (currentFloor - MinFloor) + (floor - MinFloor) + 2

	default: // Idle
		diff := floor - currentFloor
		if diff < 0 {
			return -diff
		}
		return diff
	}
}
