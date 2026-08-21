package vending_machine

import "fmt"

// State represents the common interface for all concrete states.
type State interface {
	InsertCoin() error
	EjectCoin() error
	SelectProduct() error
	Dispense() error
}

// VendingMachine is the Context in the State Pattern.
type VendingMachine struct {
	hasCoinState       State
	noCoinState        State
	itemDispensedState State
	soldOutState       State

	currentState State
	itemCount    int
}

// NewVendingMachine initializes the context and all concrete states.
func NewVendingMachine(itemCount int) *VendingMachine {
	v := &VendingMachine{
		itemCount: itemCount,
	}

	noCoinState := &NoCoinState{vendingMachine: v}
	hasCoinState := &HasCoinState{vendingMachine: v}
	itemDispensedState := &ItemDispensedState{vendingMachine: v}
	soldOutState := &SoldOutState{vendingMachine: v}

	v.hasCoinState = hasCoinState
	v.noCoinState = noCoinState
	v.itemDispensedState = itemDispensedState
	v.soldOutState = soldOutState

	if itemCount > 0 {
		v.currentState = noCoinState
	} else {
		v.currentState = soldOutState
	}

	return v
}

// Delegates operations to the current state object.

func (v *VendingMachine) InsertCoin() error {
	return v.currentState.InsertCoin()
}

func (v *VendingMachine) EjectCoin() error {
	return v.currentState.EjectCoin()
}

func (v *VendingMachine) SelectProduct() error {
	err := v.currentState.SelectProduct()
	if err != nil {
		return err
	}
	return v.currentState.Dispense()
}

func (v *VendingMachine) SetState(s State) {
	v.currentState = s
}

func (v *VendingMachine) GetHasCoinState() State {
	return v.hasCoinState
}

func (v *VendingMachine) GetNoCoinState() State {
	return v.noCoinState
}

func (v *VendingMachine) GetItemDispensedState() State {
	return v.itemDispensedState
}

func (v *VendingMachine) GetSoldOutState() State {
	return v.soldOutState
}

func (v *VendingMachine) GetItemCount() int {
	return v.itemCount
}

func (v *VendingMachine) DecrementItemCount() {
	if v.itemCount > 0 {
		v.itemCount--
	}
}

func (v *VendingMachine) DisplayCurrentState() {
	fmt.Printf("Current state: %T, Items remaining: %d\n", v.currentState, v.itemCount)
}
