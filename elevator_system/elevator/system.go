package elevator

import "sync"

// ElevatorSystem represents the central service managing the fleet of elevators.
// It acts as a Facade coordinating elevator requests, scheduling, and discrete simulation steps.
type ElevatorSystem struct {
	elevators []*Elevator
	scheduler Scheduler
	config    *Config
	mu        sync.RWMutex
}

// NewElevatorSystem initializes an elevator system with the given config and scheduler.
func NewElevatorSystem(config *Config, scheduler Scheduler) *ElevatorSystem {
	if config == nil {
		config = DefaultConfig()
	}
	if scheduler == nil {
		scheduler = NewNearestElevatorScheduler()
	}

	elevators := make([]*Elevator, config.TotalElevators)
	for i := 0; i < config.TotalElevators; i++ {
		elevators[i] = NewElevator(i, config)
	}

	return &ElevatorSystem{
		elevators: elevators,
		scheduler: scheduler,
		config:    config,
	}
}

// RequestElevator handles an external hall call from a floor with a direction.
// Fulfills Requirements 2, 5, 6, 7, 8:
// - Dispatches to the best elevator.
// - Rejects invalid floors/directions by returning false.
// - Treats current floor requests as no-op/already served.
// - Thread-safe for multiple concurrent requests.
func (s *ElevatorSystem) RequestElevator(floor int, dir Direction) bool {
	// Validate floor bounds (Requirement 7: 0-9)
	if floor < MinFloor || floor > MaxFloor {
		return false
	}

	// Validate direction
	if dir != Up && dir != Down {
		return false
	}
	if floor == MinFloor && dir == Down {
		return false
	}
	if floor == MaxFloor && dir == Up {
		return false
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Requirement 8: If an idle elevator is already at the floor, treat as no-op / already served
	for _, e := range s.elevators {
		if e.GetCurrentFloor() == floor && e.GetState() == StateIdle {
			return true
		}
	}

	// Requirement 2: System decides which elevator to dispatch
	selected := s.scheduler.SelectElevator(s.elevators, floor, dir)
	if selected == nil {
		return false
	}

	return selected.AddHallCall(floor, dir)
}

// SelectDestination handles an internal floor request from inside an elevator.
// Fulfills Requirements 3, 5, 6, 7, 8:
// - Allows selecting one or more destination floors.
// - Rejects invalid elevator IDs or floors by returning false.
// - Treats current floor requests as no-op/already served.
// - Thread-safe for multiple concurrent requests.
func (s *ElevatorSystem) SelectDestination(elevatorID int, floor int) bool {
	// Validate elevator ID bounds (Requirement 7)
	if elevatorID < 0 || elevatorID >= len(s.elevators) {
		return false
	}

	// Validate floor bounds (Requirement 7: 0-9)
	if floor < MinFloor || floor > MaxFloor {
		return false
	}

	s.mu.RLock()
	e := s.elevators[elevatorID]
	s.mu.RUnlock()

	return e.AddDestination(floor)
}

// Step advances the entire elevator system simulation by one discrete time step (Requirement 4).
func (s *ElevatorSystem) Step() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, e := range s.elevators {
		e.Step()
	}
}

// Tick is an alias for Step to advance discrete time.
func (s *ElevatorSystem) Tick() {
	s.Step()
}

// GetElevators returns the list of managed elevators.
func (s *ElevatorSystem) GetElevators() []*Elevator {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.elevators
}

// GetElevator returns a specific elevator by ID.
func (s *ElevatorSystem) GetElevator(id int) (*Elevator, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if id < 0 || id >= len(s.elevators) {
		return nil, false
	}
	return s.elevators[id], true
}
