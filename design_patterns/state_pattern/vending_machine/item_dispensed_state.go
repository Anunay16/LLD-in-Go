package vending_machine

import "fmt"

// ItemDispensedState represents the state when an item is being dispensed.
type ItemDispensedState struct {
	vendingMachine *VendingMachine
}

func (s *ItemDispensedState) InsertCoin() error {
	return fmt.Errorf("please wait: currently dispensing item")
}

func (s *ItemDispensedState) EjectCoin() error {
	return fmt.Errorf("cannot eject coin: product already selected and being dispensed")
}

func (s *ItemDispensedState) SelectProduct() error {
	return fmt.Errorf("please wait: currently dispensing item")
}

func (s *ItemDispensedState) Dispense() error {
	s.vendingMachine.DecrementItemCount()
	fmt.Println("--> Action: Dispense -> Item dispensed. Enjoy!")

	if s.vendingMachine.GetItemCount() > 0 {
		s.vendingMachine.SetState(s.vendingMachine.GetNoCoinState())
	} else {
		fmt.Println("--> System Alert: Vending machine is now out of stock!")
		s.vendingMachine.SetState(s.vendingMachine.GetSoldOutState())
	}
	return nil
}
