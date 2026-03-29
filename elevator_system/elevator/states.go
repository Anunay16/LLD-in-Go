package elevator

import "time"

type ElevatorState int

const (
	StateIdle ElevatorState = iota
	StateMovingUp
	StateMovingDown
	StateDoorOpen
)

func (s ElevatorState) String() string {
	switch s {
	case StateDoorOpen:
		return "DOOR_OPEN"
	case StateMovingUp:
		return "MOVING_UP"
	case StateMovingDown:
		return "MOVING_DOWN"
	default:
		return "IDLE"
	}
}

func (s ElevatorState) GetDirection() string {
	switch s {
	case StateMovingUp:
		return "UP"
	case StateMovingDown:
		return "DOWN"
	default:
		return "IDLE"
	}
}

type ElevatorStateHandler interface {
	// handle the current state and return the next state
	HandleState(elevator *Elevator) ElevatorStateHandler
	// Get the state name for identification
	GetState() ElevatorState
	// Enter the current state (called when transitioning to this state)
	Enter(elevator *Elevator)
	// Exit from the current state (called when leaving this state)
	Exit(elevator *Elevator)
}

// concrete states which will implement the above interface
// IdleState - Elevator is waiting for requests
type IdleState struct {
}

func (s *IdleState) HandleState(e *Elevator) ElevatorStateHandler {
	if e.queue.isEmpty() {
		return s
	}

	nextFloor, ok := e.queue.getNextFloor(Idle)
	if !ok {
		return s
	}

	if nextFloor > e.currentFloor {
		return &MovingUpState{}
	} else if nextFloor < e.currentFloor {
		return &MovingDownState{}
	}

	return &DoorOpenState{}
}

func (s *IdleState) GetState() ElevatorState {
	return StateIdle
}

func (s *IdleState) Enter(e *Elevator) {
	e.direction = Idle
}

func (s *IdleState) Exit(e *Elevator) {}

// MovingUpState - Elevator is moving upwards
type MovingUpState struct {
}

func (s *MovingUpState) HandleState(e *Elevator) ElevatorStateHandler {
	e.currentFloor++

	// check if we need to stop
	nextFloor, ok := e.queue.getNextFloor(Up)
	if ok && nextFloor == e.currentFloor {
		e.queue.removeFloor(e.currentFloor)
		return &DoorOpenState{}
	}

	// no more upward requests
	if len(e.queue.upQueue) == 0 {
		if len(e.queue.downQueue) > 0 {
			return &MovingDownState{}
		}
		return &IdleState{}
	}

	return s
}

func (s *MovingUpState) GetState() ElevatorState {
	return StateMovingUp
}

func (s *MovingUpState) Enter(e *Elevator) {
	e.direction = Up
}

func (s *MovingUpState) Exit(e *Elevator) {}

// MovingDownState - Elevator is moving downwards
type MovingDownState struct {
}

func (s *MovingDownState) HandleState(e *Elevator) ElevatorStateHandler {
	e.currentFloor--

	// check if we need to stop
	nextFloor, ok := e.queue.getNextFloor(Down)
	if ok && nextFloor == e.currentFloor {
		e.queue.removeFloor(e.currentFloor)
		return &DoorOpenState{}
	}

	// no more downward requests
	if len(e.queue.downQueue) == 0 {
		if len(e.queue.upQueue) > 0 {
			return &MovingUpState{}
		}
		return &IdleState{}
	}

	return s
}

func (s *MovingDownState) GetState() ElevatorState {
	return StateMovingDown
}

func (s *MovingDownState) Enter(e *Elevator) {
	e.direction = Down
}

func (s *MovingDownState) Exit(e *Elevator) {}

// DoorOpenState - Elevator door is open
type DoorOpenState struct {
}

func (s *DoorOpenState) HandleState(e *Elevator) ElevatorStateHandler {
	// simulate door open time
	time.Sleep(500 * time.Millisecond)

	if e.queue.isEmpty() {
		return &IdleState{}
	}

	// decide next direction based on remaining requests
	if len(e.queue.upQueue) > 0 {
		return &MovingUpState{}
	}
	if len(e.queue.downQueue) > 0 {
		return &MovingDownState{}
	}

	return &IdleState{}
}

func (s *DoorOpenState) GetState() ElevatorState {
	return StateDoorOpen
}

func (s *DoorOpenState) Enter(e *Elevator) {
	e.isDoorOpen = true
}

func (s *DoorOpenState) Exit(e *Elevator) {
	e.isDoorOpen = false
}
