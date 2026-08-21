package vending_machine

import "fmt"

// HasCoinState represents the state when a coin is inserted in the machine.
type HasCoinState struct {
	vendingMachine *VendingMachine
}

func (s *HasCoinState) InsertCoin() error {
	return fmt.Errorf("cannot insert coin: coin already inserted")
}

func (s *HasCoinState) EjectCoin() error {
	fmt.Println("--> Action: EjectCoin -> Coin returned.")
	s.vendingMachine.SetState(s.vendingMachine.GetNoCoinState())
	return nil
}

func (s *HasCoinState) SelectProduct() error {
	fmt.Println("--> Action: SelectProduct -> Product selected.")
	s.vendingMachine.SetState(s.vendingMachine.GetItemDispensedState())
	return nil
}

func (s *HasCoinState) Dispense() error {
	return fmt.Errorf("cannot dispense: select a product first")
}
