package locker

type CarrierEntryState struct {
	machine *LockerMachine
}

func (ces *CarrierEntryState) Touch() {
}

func (ces *CarrierEntryState) ValidateCode(code string) {
}

func (ces *CarrierEntryState) CloseDoor(_ *Slot) {
}

func (ces *CarrierEntryState) SelectCarrierEntry() {
}
