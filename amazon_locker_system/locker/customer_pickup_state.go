package locker

import (
	"fmt"
	"time"
)

type CustomerPickupState struct {
	machine *LockerMachine
}

func (cps *CustomerPickupState) Touch() {
}

func (cps *CustomerPickupState) ValidateCode(code string) {
	if len(code) == 0 {
		fmt.Println("Wrong code entered. Try again")
		return
	}
	fmt.Println("Door is Opened. Take out the Item")
	time.Sleep(2 * time.Second)
}

func (cps *CustomerPickupState) CloseDoor(slot *Slot) {
	fmt.Println("Door closed. Switching back to IDLE_STATE")
	fmt.Printf("slot free size: %s\n", slot.GetSize())
	cps.machine.SetState(&IdleLockerState{machine: cps.machine})
}

func (cps *CustomerPickupState) SelectCarrierEntry() {
	fmt.Println("Carrier Entry selected. Switching to CARRIER_ENTRY")
	cps.machine.SetState(&CarrierEntryState{machine: cps.machine})
}
