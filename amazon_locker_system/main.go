package main

import (
	"fmt"
	"practice/als/locker"
	lockerservice "practice/als/locker_service"
)

func main() {
	lockerService := lockerservice.NewLockerService()

	// user puts a zipcode and lockerservice returns the available lockers for the package type
	eligibleLockersForPackage, err := lockerService.GetEligibleLockersByZipAndSize("560048", locker.LargePackageSize)
	if err != nil {
		fmt.Println(err)
		return
	}

	pkg := locker.NewPackage("pkg1", "agent1", locker.LargePackageSize)

	// user choose a locker, and a slot is alloted in the backend
	_, err = lockerService.ReserveSlotInLocker(eligibleLockersForPackage[0], pkg)
	if err != nil {
		fmt.Println(err)
		return
	}

	// delivery agent
}
