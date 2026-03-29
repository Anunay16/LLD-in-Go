package locker

type Locker struct {
	slots   []*Slot
	name    string
	zipCode string
}

func NewLocker(name string, slots []*Slot, zipcode string) *Locker {
	return &Locker{
		slots:   slots,
		name:    name,
		zipCode: zipcode,
	}
}

func (l *Locker) GetName() string {
	return l.name
}

func (l *Locker) IsSlotAvailableForPackage() bool {
	for _, slot := range l.slots {
		if slot.IsAvailable() {
			return true
		}
	}
	return false
}

func (l *Locker) GetAllSlots() []*Slot {
	return l.slots
}
