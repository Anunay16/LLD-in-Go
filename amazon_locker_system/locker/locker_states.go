package locker

type LockerState interface {
	Touch()
	ValidateCode(code string)
	CloseDoor(slot *Slot)
	SelectCarrierEntry()
}
