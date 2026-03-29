package repository

import (
	"fmt"
	"practice/als/locker"
)

type LockerRepository struct {
	zipToLockers map[string][]*locker.Locker
	nameToLocker map[string]*locker.Locker
}

func NewLockerRepository() *LockerRepository {
	slot11 := locker.NewSlot(locker.LargeSlotSize)
	slot12 := locker.NewSlot(locker.MediumSlotSize)
	slot13 := locker.NewSlot(locker.SmallSlotSize)
	locker1 := locker.NewLocker("locker1", []*locker.Slot{slot11, slot12, slot13}, "560048")
	slot21 := locker.NewSlot(locker.LargeSlotSize)
	slot22 := locker.NewSlot(locker.MediumSlotSize)
	slot23 := locker.NewSlot(locker.SmallSlotSize)
	locker2 := locker.NewLocker("locker2", []*locker.Slot{slot21, slot22, slot23}, "560049")
	slot31 := locker.NewSlot(locker.LargeSlotSize)
	slot32 := locker.NewSlot(locker.MediumSlotSize)
	slot33 := locker.NewSlot(locker.SmallSlotSize)
	locker3 := locker.NewLocker("locker3", []*locker.Slot{slot31, slot32, slot33}, "560049")
	zipToLockerObj := make(map[string][]*locker.Locker)
	zipToLockerObj["560048"] = []*locker.Locker{locker1}
	zipToLockerObj["560049"] = []*locker.Locker{locker2, locker3}
	nameToLocker := make(map[string]*locker.Locker)
	nameToLocker["locker1"] = locker1
	nameToLocker["locker2"] = locker2
	nameToLocker["locker3"] = locker3
	return &LockerRepository{
		zipToLockers: zipToLockerObj,
		nameToLocker: nameToLocker,
	}
}

func (lr *LockerRepository) GetLockersByZipCode(zipCode string) ([]*locker.Locker, error) {
	lockers, ok := lr.zipToLockers[zipCode]
	if !ok {
		return nil, fmt.Errorf("no lockers found for the zipcode %s", zipCode)
	}
	return lockers, nil
}

func (lr *LockerRepository) GetLockerByName(name string) (*locker.Locker, error) {
	locker, ok := lr.nameToLocker[name]
	if !ok {
		return nil, fmt.Errorf("no locker found with the specified name")
	}
	return locker, nil
}
