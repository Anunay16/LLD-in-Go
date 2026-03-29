package locker

import (
	"fmt"
)

type IdleLockerState struct {
	machine *LockerMachine
}

func (ils *IdleLockerState) Touch() {
	fmt.Println("Screen touched -> switching to CUSTOMER_PICKUP")
	ils.machine.SetState(&CustomerPickupState{machine: ils.machine})
}

func (ils *IdleLockerState) ValidateCode(code string) {
}

func (ils *IdleLockerState) CloseDoor(_ *Slot) {
}

func (ils *IdleLockerState) SelectCarrierEntry() {
}
