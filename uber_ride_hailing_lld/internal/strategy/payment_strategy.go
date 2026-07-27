package strategy

import "fmt"

type PaymentStrategy interface {
	ProcessPayment(tripID string, amount float64) (bool, string)
}

type CashPaymentStrategy struct{}

func NewCashPaymentStrategy() *CashPaymentStrategy {
	return &CashPaymentStrategy{}
}

func (s *CashPaymentStrategy) ProcessPayment(tripID string, amount float64) (bool, string) {
	return true, fmt.Sprintf("Paid ₹%.2f in CASH for trip %s", amount, tripID)
}

type CreditCardPaymentStrategy struct {
	CardNumber string
}

func NewCreditCardPaymentStrategy(cardNumber string) *CreditCardPaymentStrategy {
	return &CreditCardPaymentStrategy{CardNumber: cardNumber}
}

func (s *CreditCardPaymentStrategy) ProcessPayment(tripID string, amount float64) (bool, string) {
	maskedCard := "****"
	if len(s.CardNumber) >= 4 {
		maskedCard = s.CardNumber[len(s.CardNumber)-4:]
	}
	return true, fmt.Sprintf("Paid ₹%.2f via CREDIT CARD (ending %s) for trip %s", amount, maskedCard, tripID)
}

type WalletPaymentStrategy struct {
	WalletID string
	Balance  float64
}

func NewWalletPaymentStrategy(walletID string, balance float64) *WalletPaymentStrategy {
	return &WalletPaymentStrategy{
		WalletID: walletID,
		Balance:  balance,
	}
}

func (s *WalletPaymentStrategy) ProcessPayment(tripID string, amount float64) (bool, string) {
	if s.Balance < amount {
		return false, fmt.Sprintf("Insufficient wallet balance (₹%.2f available, ₹%.2f required)", s.Balance, amount)
	}
	s.Balance -= amount
	return true, fmt.Sprintf("Paid ₹%.2f via WALLET (%s) for trip %s. Remaining balance: ₹%.2f", amount, s.WalletID, tripID, s.Balance)
}
