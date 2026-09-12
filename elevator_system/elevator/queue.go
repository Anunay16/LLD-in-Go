package elevator

import "sync"

// StopQueue manages the pending stops for an elevator using the LOOK/SCAN algorithm.
type StopQueue struct {
	destinations  map[int]bool // internal destination requests
	upHallCalls   map[int]bool // external hall calls going UP
	downHallCalls map[int]bool // external hall calls going DOWN
	mu            sync.RWMutex
}

func newStopQueue() *StopQueue {
	return &StopQueue{
		destinations:  make(map[int]bool),
		upHallCalls:   make(map[int]bool),
		downHallCalls: make(map[int]bool),
	}
}

// AddDestination adds an internal destination floor request.
func (q *StopQueue) AddDestination(floor int) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.destinations[floor] = true
}

// AddHallCall adds an external hall call request with direction.
func (q *StopQueue) AddHallCall(floor int, dir Direction) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if dir == Up {
		q.upHallCalls[floor] = true
	} else if dir == Down {
		q.downHallCalls[floor] = true
	}
}

// HasStops returns true if there are any pending stops.
func (q *StopQueue) HasStops() bool {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.destinations) > 0 || len(q.upHallCalls) > 0 || len(q.downHallCalls) > 0
}

// HasStopsAbove returns true if any requested stop is above the given floor.
func (q *StopQueue) HasStopsAbove(floor int) bool {
	for f := range q.destinations {
		if f > floor {
			return true
		}
	}
	for f := range q.upHallCalls {
		if f > floor {
			return true
		}
	}
	for f := range q.downHallCalls {
		if f > floor {
			return true
		}
	}
	return false
}

// HasStopsBelow returns true if any requested stop is below the given floor.
func (q *StopQueue) HasStopsBelow(floor int) bool {
	for f := range q.destinations {
		if f < floor {
			return true
		}
	}
	for f := range q.upHallCalls {
		if f < floor {
			return true
		}
	}
	for f := range q.downHallCalls {
		if f < floor {
			return true
		}
	}
	return false
}

// ShouldStop checks whether the elevator should stop at the given floor given its current direction.
func (q *StopQueue) ShouldStop(floor int, dir Direction) bool {
	q.mu.RLock()
	defer q.mu.RUnlock()

	// Internal destinations always stop
	if q.destinations[floor] {
		return true
	}

	switch dir {
	case Up:
		if q.upHallCalls[floor] {
			return true
		}
		// If at top turnaround and a down call exists at this floor
		if q.downHallCalls[floor] && !q.HasStopsAbove(floor) {
			return true
		}
	case Down:
		if q.downHallCalls[floor] {
			return true
		}
		// If at bottom turnaround and an up call exists at this floor
		if q.upHallCalls[floor] && !q.HasStopsBelow(floor) {
			return true
		}
	case Idle:
		if q.upHallCalls[floor] || q.downHallCalls[floor] {
			return true
		}
	}

	return false
}

// ClearStopsAt clears the stops served at the given floor.
func (q *StopQueue) ClearStopsAt(floor int, dir Direction) {
	q.mu.Lock()
	defer q.mu.Unlock()

	delete(q.destinations, floor)

	switch dir {
	case Up:
		delete(q.upHallCalls, floor)
		if !q.HasStopsAbove(floor) {
			delete(q.downHallCalls, floor)
		}
	case Down:
		delete(q.downHallCalls, floor)
		if !q.HasStopsBelow(floor) {
			delete(q.upHallCalls, floor)
		}
	case Idle:
		delete(q.upHallCalls, floor)
		delete(q.downHallCalls, floor)
	}
}

// DetermineNextDirection decides the direction the elevator should head in next.
func (q *StopQueue) DetermineNextDirection(currentFloor int, currentDir Direction) Direction {
	q.mu.RLock()
	defer q.mu.RUnlock()

	switch currentDir {
	case Up:
		if q.HasStopsAbove(currentFloor) {
			return Up
		}
		if q.HasStopsBelow(currentFloor) {
			return Down
		}
		return Idle
	case Down:
		if q.HasStopsBelow(currentFloor) {
			return Down
		}
		if q.HasStopsAbove(currentFloor) {
			return Up
		}
		return Idle
	default: // Idle
		if q.HasStopsAbove(currentFloor) {
			return Up
		}
		if q.HasStopsBelow(currentFloor) {
			return Down
		}
		return Idle
	}
}
