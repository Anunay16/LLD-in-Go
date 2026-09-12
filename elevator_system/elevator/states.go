package elevator

// ElevatorStateHandler represents the State Pattern interface for elevator behavior.
type ElevatorStateHandler interface {
	HandleStep(e *Elevator) ElevatorStateHandler
	GetState() ElevatorState
	GetDirection() Direction
}

// IdleState represents an elevator waiting for requests.
type IdleState struct{}

func (s *IdleState) GetState() ElevatorState {
	return StateIdle
}

func (s *IdleState) GetDirection() Direction {
	return Idle
}

func (s *IdleState) HandleStep(e *Elevator) ElevatorStateHandler {
	if !e.queue.HasStops() {
		return s
	}

	// Check if there are stops right at the current floor
	if e.queue.ShouldStop(e.currentFloor, Idle) {
		e.queue.ClearStopsAt(e.currentFloor, Idle)
		e.notifyFloorServed(e.currentFloor)
	}

	nextDir := e.queue.DetermineNextDirection(e.currentFloor, Idle)
	switch nextDir {
	case Up:
		e.currentFloor++
		e.notifyFloorChanged(e.currentFloor - 1, e.currentFloor)
		if e.queue.ShouldStop(e.currentFloor, Up) {
			e.queue.ClearStopsAt(e.currentFloor, Up)
			e.notifyFloorServed(e.currentFloor)
		}
		remainingDir := e.queue.DetermineNextDirection(e.currentFloor, Up)
		switch remainingDir {
		case Up:
			return &MovingUpState{}
		case Down:
			return &MovingDownState{}
		default:
			return &IdleState{}
		}
	case Down:
		e.currentFloor--
		e.notifyFloorChanged(e.currentFloor + 1, e.currentFloor)
		if e.queue.ShouldStop(e.currentFloor, Down) {
			e.queue.ClearStopsAt(e.currentFloor, Down)
			e.notifyFloorServed(e.currentFloor)
		}
		remainingDir := e.queue.DetermineNextDirection(e.currentFloor, Down)
		switch remainingDir {
		case Down:
			return &MovingDownState{}
		case Up:
			return &MovingUpState{}
		default:
			return &IdleState{}
		}
	default:
		return s
	}
}

// MovingUpState represents an elevator moving towards higher floors.
type MovingUpState struct{}

func (s *MovingUpState) GetState() ElevatorState {
	return StateMovingUp
}

func (s *MovingUpState) GetDirection() Direction {
	return Up
}

func (s *MovingUpState) HandleStep(e *Elevator) ElevatorStateHandler {
	oldFloor := e.currentFloor
	e.currentFloor++
	e.notifyFloorChanged(oldFloor, e.currentFloor)

	if e.queue.ShouldStop(e.currentFloor, Up) {
		e.queue.ClearStopsAt(e.currentFloor, Up)
		e.notifyFloorServed(e.currentFloor)
	}

	nextDir := e.queue.DetermineNextDirection(e.currentFloor, Up)
	switch nextDir {
	case Up:
		return s
	case Down:
		return &MovingDownState{}
	default:
		return &IdleState{}
	}
}

// MovingDownState represents an elevator moving towards lower floors.
type MovingDownState struct{}

func (s *MovingDownState) GetState() ElevatorState {
	return StateMovingDown
}

func (s *MovingDownState) GetDirection() Direction {
	return Down
}

func (s *MovingDownState) HandleStep(e *Elevator) ElevatorStateHandler {
	oldFloor := e.currentFloor
	e.currentFloor--
	e.notifyFloorChanged(oldFloor, e.currentFloor)

	if e.queue.ShouldStop(e.currentFloor, Down) {
		e.queue.ClearStopsAt(e.currentFloor, Down)
		e.notifyFloorServed(e.currentFloor)
	}

	nextDir := e.queue.DetermineNextDirection(e.currentFloor, Down)
	switch nextDir {
	case Down:
		return s
	case Up:
		return &MovingUpState{}
	default:
		return &IdleState{}
	}
}
