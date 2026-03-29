package elevator

import (
	"fmt"
	"sync"
	"time"
)

type Elevator struct {
	id           int
	currentFloor int
	stateHandler ElevatorStateHandler
	isDoorOpen   bool
	direction    Direction

	// request management
	queue *elevatorQueue

	config    *Config
	observers []ElevatorObserver
	prevState ElevatorState
	prevFloor int

	// synchronization
	mu          sync.RWMutex
	stopChan    chan struct{}
	requestChan chan *Request
}

func NewElevator(id int, config *Config) *Elevator {
	elevator := &Elevator{
		id:           id,
		currentFloor: 0,
		stateHandler: &IdleState{},
		isDoorOpen:   false,
		direction:    Idle,
		queue:        newElevatorQueue(),
		config:       config,
		observers:    make([]ElevatorObserver, 0),
		prevState:    StateIdle,
		prevFloor:    0,
		stopChan:     make(chan struct{}),
		requestChan:  make(chan *Request, 100), // Buffered channel
	}
	elevator.stateHandler.Enter(elevator)
	return elevator
}

func (e *Elevator) AddRequest(request *Request) error {
	// validate request
	if request == nil {
		return fmt.Errorf("request can not be nil")
	}
	if request.FromFloor < MinFloors || request.FromFloor > e.config.TotalFloors ||
		request.ToFloor < MinFloors || request.ToFloor > e.config.TotalFloors {
		return fmt.Errorf("invalid floor numbers: from=%d, to=%d", request.FromFloor, request.ToFloor)
	}

	// make it safe for many goroutines
	select {
	case e.requestChan <- request:
		return nil
	default:
		return fmt.Errorf("elevator %d request queue is full", e.id)
	}
}

func (e *Elevator) GetState() ElevatorState {
	return e.stateHandler.GetState()
}

func (e *Elevator) GetCurrentFloor() int {
	return e.currentFloor
}

func (e *Elevator) Start() {
	go e.run()
}

func (e *Elevator) Stop() {
	close(e.stopChan)
}

// run is the main loop for elevator using state pattern with integrated observer notification
func (e *Elevator) run() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-e.stopChan:
			e.TransitionToState(&IdleState{})
			e.NotifyObservers()
			return

		case request := <-e.requestChan:
			e.processRequest(request)

		case <-ticker.C:
			e.processMovement()
		}

		// notify after every meaningful change
		e.NotifyObservers()
	}
}

func (e *Elevator) TransitionToState(newState ElevatorStateHandler) {
	if e.stateHandler != nil {
		e.stateHandler.Exit(e)
	}
	e.stateHandler = newState
	if newState != nil {
		e.stateHandler.Enter(e)
	}
}

func (e *Elevator) NotifyObservers() {
	currentState := e.GetState()
	currentFloor := e.GetCurrentFloor()

	// Notify state change
	if currentState != e.prevState {
		for _, observer := range e.observers {
			observer.OnElevatorStateChanged(e.id, e.prevState, currentState)
		}
		e.prevState = currentState
	}

	// Notify floor change
	if currentFloor != e.prevFloor {
		for _, observer := range e.observers {
			observer.OnElevatorFloorChanged(e.id, e.prevFloor, currentFloor)
		}
		e.prevFloor = currentFloor
	}
}

func (e *Elevator) processRequest(request *Request) {
	e.mu.Lock()
	defer e.mu.Unlock()

	currentFloor := e.currentFloor

	// External request → first go to pickup floor
	if request.Type == RequestTypeExternal {
		if request.FromFloor != currentFloor {
			e.queue.addFloor(request.FromFloor, currentFloor)
		}
	}

	// Always add destination floor
	if request.ToFloor != currentFloor {
		e.queue.addFloor(request.ToFloor, currentFloor)
	}
}

func (e *Elevator) processMovement() {
	e.mu.Lock()
	defer e.mu.Unlock()

	nextState := e.stateHandler.HandleState(e)
	if nextState != e.stateHandler {
		e.TransitionToState(nextState)
	}
}
