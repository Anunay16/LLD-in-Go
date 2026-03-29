package lockerservice

import (
	"fmt"
	"practice/als/locker"
)

type FirstFitSlotStrategy struct {
}

func (ff *FirstFitSlotStrategy) assignSlot(eligibleSlots []*locker.Slot) (*locker.Slot, error) {
	for _, slot := range eligibleSlots {
		if slot.Acquire() {
			return slot, nil
		}
	}
	return nil, fmt.Errorf("no availble slots found in the locker")
}
