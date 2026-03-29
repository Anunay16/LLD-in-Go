package elevator

const (
	MinFloors             = 1
	DefaultTotalFloors    = 10
	DefaultTotalElevators = 3
)

type Config struct {
	TotalFloors    int
	TotalElevators int
	SchedulerType  string
}

func DefaultConfig() *Config {
	return &Config{
		TotalFloors:    DefaultTotalFloors,
		TotalElevators: DefaultTotalElevators,
		SchedulerType:  "SCAN",
	}
}

// ElevatorObserver is the interface for observing elevator events
// Observer: ElevatorController, This should be implemented by the ElevatorController to monitor all elevators
// Observable: Elevator
type ElevatorObserver interface {
	OnElevatorStateChanged(elevatorId int, oldState, newState ElevatorState)
	OnElevatorFloorChanged(elevatorId int, oldFloor, newFloor int)
	OnRequestComplete(elevatorId int, requestId string)
}

type RequestType int

const (
	RequestTypeExternal RequestType = iota // called from outside elevator
	RequestTypeInternal                    // called from inside elevator
)

type Direction int

const (
	Up Direction = iota
	Down
	Idle
)

type Request struct {
	ID        int
	FromFloor int
	ToFloor   int
	Type      RequestType
	Direction Direction
}
