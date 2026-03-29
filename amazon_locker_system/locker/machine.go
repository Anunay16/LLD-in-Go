package locker

type LockerMachine struct {
	locker *Locker
	state  LockerState
}

func NewLockerMachine(lc *Locker) *LockerMachine {
	machine := &LockerMachine{
		locker: lc,
	}
	idleState := &IdleLockerState{
		machine: machine,
	}
	machine.SetState(idleState)
	return machine
}

func (lm *LockerMachine) Touch() {
	lm.state.Touch()
}

func (lm *LockerMachine) ValidateCode(code string) {
	lm.state.ValidateCode(code)
}

func (lm *LockerMachine) CloseDoor(slot *Slot) {
	lm.state.CloseDoor(slot)
}

func (lm *LockerMachine) SetState(state LockerState) {
	lm.state = state
}
