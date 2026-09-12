package elevator

import (
	"sync"
)

// Elevator represents a single elevator instance.
type Elevator struct {
	id           int
	currentFloor int
	stateHandler ElevatorStateHandler
	queue        *StopQueue
	config       *Config
	observers    []ElevatorObserver
	mu           sync.RWMutex
}

// NewElevator initializes an elevator at floor 0 in Idle state.
func NewElevator(id int, config *Config) *Elevator {
	if config == nil {
		config = DefaultConfig()
	}
	return &Elevator{
		id:           id,
		currentFloor: MinFloor,
		stateHandler: &IdleState{},
		queue:        newStopQueue(),
		config:       config,
		observers:    make([]ElevatorObserver, 0),
	}
}

// GetID returns the elevator identifier.
func (e *Elevator) GetID() int {
	return e.id
}

// GetCurrentFloor returns the current floor of the elevator.
func (e *Elevator) GetCurrentFloor() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.currentFloor
}

// GetState returns the current operational state of the elevator.
func (e *Elevator) GetState() ElevatorState {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.stateHandler.GetState()
}

// GetDirection returns the current moving direction or intended direction.
func (e *Elevator) GetDirection() Direction {
	e.mu.RLock()
	defer e.mu.RUnlock()
	dir := e.stateHandler.GetDirection()
	if dir == Idle && e.queue.HasStops() {
		return e.queue.DetermineNextDirection(e.currentFloor, Idle)
	}
	return dir
}

// HasStops returns true if the elevator has pending stops.
func (e *Elevator) HasStops() bool {
	return e.queue.HasStops()
}

// AddDestination registers an internal destination request.
// Returns false if floor is out of range.
// Returns true if successfully registered or if already at the requested floor (no-op).
func (e *Elevator) AddDestination(floor int) bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	if floor < MinFloor || floor > MaxFloor {
		return false
	}

	// Requirement 8: Requests for current floor are treated as a no-op / already served
	if floor == e.currentFloor {
		return true
	}

	e.queue.AddDestination(floor)
	return true
}

// AddHallCall registers an external pickup request with direction.
// Returns false if floor or direction is invalid.
// Returns true if successfully registered or if already at the requested floor (no-op).
func (e *Elevator) AddHallCall(floor int, dir Direction) bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	if floor < MinFloor || floor > MaxFloor {
		return false
	}
	if dir != Up && dir != Down {
		return false
	}
	if floor == MinFloor && dir == Down {
		return false
	}
	if floor == MaxFloor && dir == Up {
		return false
	}

	// Requirement 8: Requests for current floor are treated as a no-op / already served
	if floor == e.currentFloor && (e.stateHandler.GetState() == StateIdle || e.stateHandler.GetDirection() == dir) {
		return true
	}

	e.queue.AddHallCall(floor, dir)
	return true
}

// Step advances the elevator simulation by one discrete time step.
func (e *Elevator) Step() {
	e.mu.Lock()
	defer e.mu.Unlock()

	oldState := e.stateHandler.GetState()
	nextState := e.stateHandler.HandleStep(e)

	if nextState.GetState() != oldState {
		e.stateHandler = nextState
		e.notifyStateChanged(oldState, nextState.GetState())
	}
}

// AddObserver registers an observer for elevator events.
func (e *Elevator) AddObserver(obs ElevatorObserver) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.observers = append(e.observers, obs)
}

func (e *Elevator) notifyStateChanged(oldState, newState ElevatorState) {
	for _, obs := range e.observers {
		obs.OnElevatorStateChanged(e.id, oldState, newState)
	}
}

func (e *Elevator) notifyFloorChanged(oldFloor, newFloor int) {
	for _, obs := range e.observers {
		obs.OnElevatorFloorChanged(e.id, oldFloor, newFloor)
	}
}

func (e *Elevator) notifyFloorServed(floor int) {
	for _, obs := range e.observers {
		obs.OnFloorServed(e.id, floor)
	}
}
