package vending_machine

import "fmt"

// SoldOutState represents the state when the machine has no inventory left.
type SoldOutState struct {
	vendingMachine *VendingMachine
}

func (s *SoldOutState) InsertCoin() error {
	return fmt.Errorf("cannot insert coin: machine is out of stock")
}

func (s *SoldOutState) EjectCoin() error {
	return fmt.Errorf("cannot eject coin: no coin was inserted and machine is out of stock")
}

func (s *SoldOutState) SelectProduct() error {
	return fmt.Errorf("cannot select product: machine is out of stock")
}

func (s *SoldOutState) Dispense() error {
	return fmt.Errorf("cannot dispense: machine is out of stock")
}
