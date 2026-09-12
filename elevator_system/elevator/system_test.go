package elevator_test

import (
	"sync"
	"testing"

	"elevator_system/elevator"
)

// Requirement 1: System manages 3 elevators serving 10 floors (0-9)
func TestSystemInitialization(t *testing.T) {
	sys := elevator.NewElevatorSystem(nil, nil)
	elevators := sys.GetElevators()

	if len(elevators) != elevator.DefaultTotalElevators {
		t.Fatalf("expected %d elevators, got %d", elevator.DefaultTotalElevators, len(elevators))
	}

	for _, e := range elevators {
		if e.GetCurrentFloor() != elevator.MinFloor {
			t.Errorf("elevator %d: expected start floor %d, got %d", e.GetID(), elevator.MinFloor, e.GetCurrentFloor())
		}
		if e.GetState() != elevator.StateIdle {
			t.Errorf("elevator %d: expected idle state, got %v", e.GetID(), e.GetState())
		}
	}
}

// Requirement 2: Users can request an elevator from any floor (hall call). System decides which elevator to dispatch.
func TestHallCallDispatch(t *testing.T) {
	sys := elevator.NewElevatorSystem(nil, nil)

	// Hall call from floor 3 going Up
	ok := sys.RequestElevator(3, elevator.Up)
	if !ok {
		t.Fatalf("expected RequestElevator to return true")
	}

	// At least one elevator should have received the call
	var dispatched *elevator.Elevator
	for _, e := range sys.GetElevators() {
		if e.HasStops() {
			dispatched = e
			break
		}
	}

	if dispatched == nil {
		t.Fatalf("expected an elevator to be dispatched")
	}

	// Step simulation until it reaches floor 3
	for i := 0; i < 3; i++ {
		sys.Step()
	}

	if dispatched.GetCurrentFloor() != 3 {
		t.Errorf("expected elevator to reach floor 3, got %d", dispatched.GetCurrentFloor())
	}
	if dispatched.HasStops() {
		t.Errorf("expected stops to be cleared at floor 3")
	}
}

// Requirement 3: Once inside, users can select one or more destination floors
func TestMultipleDestinationSelection(t *testing.T) {
	sys := elevator.NewElevatorSystem(nil, nil)

	// Inside elevator 0, passenger selects floors 2, 4, and 5
	if !sys.SelectDestination(0, 2) {
		t.Errorf("failed to select destination floor 2")
	}
	if !sys.SelectDestination(0, 4) {
		t.Errorf("failed to select destination floor 4")
	}
	if !sys.SelectDestination(0, 5) {
		t.Errorf("failed to select destination floor 5")
	}

	e, _ := sys.GetElevator(0)

	// Step 1 -> floor 1
	sys.Step()
	if e.GetCurrentFloor() != 1 {
		t.Errorf("step 1: expected floor 1, got %d", e.GetCurrentFloor())
	}

	// Step 2 -> floor 2 (first stop served)
	sys.Step()
	if e.GetCurrentFloor() != 2 {
		t.Errorf("step 2: expected floor 2, got %d", e.GetCurrentFloor())
	}

	// Step 3 -> floor 3
	sys.Step()
	if e.GetCurrentFloor() != 3 {
		t.Errorf("step 3: expected floor 3, got %d", e.GetCurrentFloor())
	}

	// Step 4 -> floor 4 (second stop served)
	sys.Step()
	if e.GetCurrentFloor() != 4 {
		t.Errorf("step 4: expected floor 4, got %d", e.GetCurrentFloor())
	}

	// Step 5 -> floor 5 (final stop served)
	sys.Step()
	if e.GetCurrentFloor() != 5 {
		t.Errorf("step 5: expected floor 5, got %d", e.GetCurrentFloor())
	}

	// Should now be Idle with no stops
	if e.HasStops() {
		t.Errorf("expected no remaining stops")
	}
	if e.GetState() != elevator.StateIdle {
		t.Errorf("expected StateIdle, got %v", e.GetState())
	}
}

// Requirement 4: Simulation runs in discrete time steps (e.g. step() or tick() call advances time)
func TestDiscreteSimulationStepAndTick(t *testing.T) {
	sys := elevator.NewElevatorSystem(nil, nil)

	sys.SelectDestination(0, 3)
	e, _ := sys.GetElevator(0)

	if e.GetCurrentFloor() != 0 {
		t.Fatalf("expected floor 0 before step, got %d", e.GetCurrentFloor())
	}

	// Step advances time deterministically
	sys.Step()
	if e.GetCurrentFloor() != 1 {
		t.Fatalf("expected floor 1 after 1 step, got %d", e.GetCurrentFloor())
	}

	// Tick is alias for Step
	sys.Tick()
	if e.GetCurrentFloor() != 2 {
		t.Fatalf("expected floor 2 after 1 tick, got %d", e.GetCurrentFloor())
	}

	sys.Step()
	if e.GetCurrentFloor() != 3 {
		t.Fatalf("expected floor 3 after 3 steps, got %d", e.GetCurrentFloor())
	}
}

// Requirement 5: Elevator stops come in two types: Hall calls (with direction) and Destination calls (no direction)
func TestTwoStopTypes(t *testing.T) {
	sys := elevator.NewElevatorSystem(nil, nil)
	e, _ := sys.GetElevator(0)

	// Elevator 0 has destination at 5
	sys.SelectDestination(0, 5)

	// Also hall call at floor 2 with UP direction should be picked up on the way
	e.AddHallCall(2, elevator.Up)

	// Step 1 -> floor 1
	sys.Step()
	// Step 2 -> floor 2 (hall call served)
	sys.Step()
	if e.GetCurrentFloor() != 2 {
		t.Errorf("expected floor 2, got %d", e.GetCurrentFloor())
	}

	// Steps to reach 5
	sys.Step() // 3
	sys.Step() // 4
	sys.Step() // 5 (destination served)
	if e.GetCurrentFloor() != 5 {
		t.Errorf("expected floor 5, got %d", e.GetCurrentFloor())
	}
	if e.HasStops() {
		t.Errorf("expected all stops to be served")
	}
}

// Requirement 6: System handles multiple concurrent pickup requests across floors
func TestConcurrentRequests(t *testing.T) {
	sys := elevator.NewElevatorSystem(nil, nil)
	var wg sync.WaitGroup

	// Launch 50 concurrent hall calls and destination requests
	for i := 0; i < 50; i++ {
		wg.Add(2)

		go func(floor int) {
			defer wg.Done()
			dir := elevator.Up
			if floor == elevator.MaxFloor {
				dir = elevator.Down
			}
			sys.RequestElevator(floor%elevator.TotalFloors, dir)
		}(i)

		go func(idx int) {
			defer wg.Done()
			elevatorID := idx % elevator.DefaultTotalElevators
			floor := (idx * 2) % elevator.TotalFloors
			sys.SelectDestination(elevatorID, floor)
		}(i)
	}

	wg.Wait()

	// Simulation should step smoothly without panic or race condition
	for step := 0; step < 20; step++ {
		sys.Step()
	}
}

// Requirement 7: Invalid requests should be rejected (return false)
// - Non-existent floor numbers (<0 or >9)
// - Non-existent elevator ID
// - Invalid hall directions (e.g. DOWN from floor 0, UP from floor 9)
func TestInvalidRequestsRejected(t *testing.T) {
	sys := elevator.NewElevatorSystem(nil, nil)

	// Negative floor
	if sys.RequestElevator(-1, elevator.Up) {
		t.Errorf("expected RequestElevator(-1, Up) to return false")
	}
	if sys.SelectDestination(0, -1) {
		t.Errorf("expected SelectDestination(0, -1) to return false")
	}

	// Floor >= 10
	if sys.RequestElevator(10, elevator.Down) {
		t.Errorf("expected RequestElevator(10, Down) to return false")
	}
	if sys.SelectDestination(0, 10) {
		t.Errorf("expected SelectDestination(0, 10) to return false")
	}

	// Invalid elevator ID
	if sys.SelectDestination(-1, 5) {
		t.Errorf("expected SelectDestination(-1, 5) to return false")
	}
	if sys.SelectDestination(3, 5) {
		t.Errorf("expected SelectDestination(3, 5) to return false")
	}

	// Invalid direction requests: DOWN from 0, UP from 9
	if sys.RequestElevator(0, elevator.Down) {
		t.Errorf("expected RequestElevator(0, Down) to return false")
	}
	if sys.RequestElevator(9, elevator.Up) {
		t.Errorf("expected RequestElevator(9, Up) to return false")
	}

	// Invalid direction enum
	if sys.RequestElevator(5, elevator.Idle) {
		t.Errorf("expected RequestElevator(5, Idle) to return false")
	}
}

// Requirement 8: Requests for the current floor are treated as a no-op / already served
func TestCurrentFloorNoOp(t *testing.T) {
	sys := elevator.NewElevatorSystem(nil, nil)
	e, _ := sys.GetElevator(0)

	// Elevator 0 starts at floor 0
	if e.GetCurrentFloor() != 0 {
		t.Fatalf("expected floor 0, got %d", e.GetCurrentFloor())
	}

	// Requesting floor 0 from inside elevator 0 should return true (treated as served/no-op)
	ok := sys.SelectDestination(0, 0)
	if !ok {
		t.Errorf("expected SelectDestination(0, 0) to return true")
	}

	// Elevator should not have any pending stops and remain Idle
	if e.HasStops() {
		t.Errorf("expected no stops added for current floor")
	}
	if e.GetState() != elevator.StateIdle {
		t.Errorf("expected StateIdle, got %v", e.GetState())
	}

	// Hall call for floor 0 when an idle elevator is already at floor 0
	okHall := sys.RequestElevator(0, elevator.Up)
	if !okHall {
		t.Errorf("expected RequestElevator(0, Up) to return true")
	}

	// Still no movement needed, already served
	sys.Step()
	if e.GetCurrentFloor() != 0 {
		t.Errorf("expected elevator to remain at floor 0, got %d", e.GetCurrentFloor())
	}
}
