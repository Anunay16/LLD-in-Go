package elevator

import (
	"sort"
	"sync"
)

type elevatorQueue struct {
	upQueue   []int // sorted ascending
	downQueue []int // sorted descending
	mu        sync.RWMutex
}

func newElevatorQueue() *elevatorQueue {
	return &elevatorQueue{
		upQueue:   make([]int, 0),
		downQueue: make([]int, 0),
	}
}

func (eq *elevatorQueue) addFloor(floor int, currentFloor int) {
	eq.mu.Lock()
	defer eq.mu.Unlock()

	// check duplicates
	if eq.exists(floor) {
		return
	}

	if floor > currentFloor {
		eq.upQueue = append(eq.upQueue, floor)
		sort.Ints(eq.upQueue) // ascending
	} else if floor < currentFloor {
		eq.downQueue = append(eq.downQueue, floor)
		sort.Sort(sort.Reverse(sort.IntSlice(eq.downQueue))) // descending
	} else {
		// same floor → can be treated as immediate stop
	}
}

func (eq *elevatorQueue) removeFloor(floor int) {
	eq.mu.Lock()
	defer eq.mu.Unlock()

	eq.upQueue = removeFromSlice(eq.upQueue, floor)
	eq.downQueue = removeFromSlice(eq.downQueue, floor)
}

func removeFromSlice(arr []int, target int) []int {
	for i, v := range arr {
		if v == target {
			return append(arr[:i], arr[i+1:]...)
		}
	}
	return arr
}

func (eq *elevatorQueue) exists(floor int) bool {
	for _, f := range eq.upQueue {
		if f == floor {
			return true
		}
	}
	for _, f := range eq.downQueue {
		if f == floor {
			return true
		}
	}
	return false
}

func (eq *elevatorQueue) isEmpty() bool {
	return len(eq.upQueue) == 0 && len(eq.downQueue) == 0
}

func (eq *elevatorQueue) getNextFloor(direction Direction) (int, bool) {
	eq.mu.RLock()
	defer eq.mu.RUnlock()

	switch direction {
	case Up:
		if len(eq.upQueue) > 0 {
			return eq.upQueue[0], true
		}
		if len(eq.downQueue) > 0 {
			return eq.downQueue[0], true
		}
	case Down:
		if len(eq.downQueue) > 0 {
			return eq.downQueue[0], true
		}
		if len(eq.upQueue) > 0 {
			return eq.upQueue[0], true
		}
	case Idle:
		// pick nearest (simple strategy)
		if len(eq.upQueue) > 0 {
			return eq.upQueue[0], true
		}
		if len(eq.downQueue) > 0 {
			return eq.downQueue[0], true
		}
	}

	return 0, false
}
