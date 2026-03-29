package elevator

type Scheduler interface {
	SelectElevator(req *Request, elevators []*Elevator) *Elevator
}

