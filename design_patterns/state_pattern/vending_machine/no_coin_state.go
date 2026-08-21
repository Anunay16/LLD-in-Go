package vending_machine

import "fmt"

// NoCoinState represents the state when no coin is inserted in the machine.
type NoCoinState struct {
	vendingMachine *VendingMachine
}

func (s *NoCoinState) InsertCoin() error {
	fmt.Println("--> Action: InsertCoin -> Coin inserted successfully.")
	s.vendingMachine.SetState(s.vendingMachine.GetHasCoinState())
	return nil
}

func (s *NoCoinState) EjectCoin() error {
	return fmt.Errorf("cannot eject coin: no coin inserted")
}

func (s *NoCoinState) SelectProduct() error {
	return fmt.Errorf("cannot select product: please insert a coin first")
}

func (s *NoCoinState) Dispense() error {
	return fmt.Errorf("cannot dispense: no coin inserted")
}
