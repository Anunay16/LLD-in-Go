package lockerservice

import (
	"fmt"
	"practice/als/locker"
	"practice/als/repository"
)

type LockerRepository interface {
	GetLockersByZipCode(zipCode string) ([]*locker.Locker, error)
	GetLockerByName(name string) (*locker.Locker, error)
}

type SlotStrategy interface {
	assignSlot(eligibleSlots []*locker.Slot) (*locker.Slot, error)
}

type LockerService struct {
	lockerRepo   LockerRepository
	slotStrategy SlotStrategy
}

func NewLockerService() *LockerService {
	return &LockerService{
		lockerRepo:   repository.NewLockerRepository(),
		slotStrategy: &FirstFitSlotStrategy{},
	}
}

func (ls *LockerService) GetEligibleLockersByZipAndSize(zipCode string, pkgSize locker.PackageSize) ([]string, error) {
	lockers, err := ls.lockerRepo.GetLockersByZipCode(zipCode)
	if err != nil {
		return nil, err
	}
	var lockerNames []string
	for _, locker := range lockers {
		if locker.IsSlotAvailableForPackage() {
			lockerNames = append(lockerNames, locker.GetName())
		}
	}
	if len(lockerNames) == 0 {
		return nil, fmt.Errorf("no eligible slots found in the locker")
	}
	return lockerNames, nil
}

func (ls *LockerService) ReserveSlotInLocker(lockerName string, pkg *locker.Package) (string, error) {
	locker, err := ls.lockerRepo.GetLockerByName(lockerName)
	if err != nil {
		return "", err
	}
	eligibleSlots := GetEligibleSlots(locker, pkg)
	if len(eligibleSlots) == 0 {
		return "", fmt.Errorf("no eligible slots found to reserve in locker %s", lockerName)
	}
	slot, err := ls.slotStrategy.assignSlot(eligibleSlots)
	if err != nil {
		return "", err
	}
	fmt.Printf("slot with id: %s is reserved for the package witd id: %s\n", slot.GetID(), pkg.GetID())
	return slot.GetID(), nil
}

func GetEligibleSlots(locker *locker.Locker, pkg *locker.Package) []*locker.Slot {
	slots := locker.GetAllSlots()
	for _, slot := range slots {
		if canFit(slot.GetSize(), pkg.GetSize()) && slot.IsAvailable() {
			slots = append(slots, slot)
		}
	}
	return slots
}

func canFit(slot locker.SlotSize, pkg locker.PackageSize) bool {
	return slot == locker.SlotSize(pkg)
}
