package elevator

const (
	MinFloor              = 0
	MaxFloor              = 9
	TotalFloors           = 10
	DefaultTotalElevators = 3
)

type Direction int

const (
	Idle Direction = iota
	Up
	Down
)

func (d Direction) String() string {
	switch d {
	case Up:
		return "UP"
	case Down:
		return "DOWN"
	default:
		return "IDLE"
	}
}

type ElevatorState int

const (
	StateIdle ElevatorState = iota
	StateMovingUp
	StateMovingDown
)

func (s ElevatorState) String() string {
	switch s {
	case StateMovingUp:
		return "MOVING_UP"
	case StateMovingDown:
		return "MOVING_DOWN"
	default:
		return "IDLE"
	}
}

type Config struct {
	TotalFloors    int
	TotalElevators int
}

func DefaultConfig() *Config {
	return &Config{
		TotalFloors:    TotalFloors,
		TotalElevators: DefaultTotalElevators,
	}
}

// ElevatorObserver is the interface for observing elevator events
type ElevatorObserver interface {
	OnElevatorStateChanged(elevatorID int, oldState, newState ElevatorState)
	OnElevatorFloorChanged(elevatorID int, oldFloor, newFloor int)
	OnFloorServed(elevatorID int, floor int)
}
