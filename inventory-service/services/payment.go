package services

import (
	"fmt"
	"strings"

	"pharmacy-order-management/models"
)

// PaymentStrategy defines the pluggable payment interface (Strategy Pattern / Open-Closed)
type PaymentStrategy interface {
	Pay(amount float64, details models.PaymentDetails) error
}

type UPIPayment struct{}

func (u *UPIPayment) Pay(amount float64, details models.PaymentDetails) error {
	if details.UPIID == "" || !strings.Contains(details.UPIID, "@") {
		return fmt.Errorf("%w: invalid UPI ID", models.ErrPaymentFailed)
	}
	fmt.Printf("[Payment] Paid $%.2f via UPI (%s)\n", amount, details.UPIID)
	return nil
}

type CreditCardPayment struct{}

func (c *CreditCardPayment) Pay(amount float64, details models.PaymentDetails) error {
	if details.CardNumber == "" || len(details.CVV) < 3 {
		return fmt.Errorf("%w: invalid card details", models.ErrPaymentFailed)
	}
	fmt.Printf("[Payment] Paid $%.2f via Credit Card (ending in %s)\n", amount, details.CardNumber[len(details.CardNumber)-4:])
	return nil
}

type CashOnDeliveryPayment struct{}

func (cod *CashOnDeliveryPayment) Pay(amount float64, details models.PaymentDetails) error {
	fmt.Printf("[Payment] Order placed with Cash on Delivery for $%.2f\n", amount)
	return nil
}

// PaymentFactory creates the appropriate payment strategy
func GetPaymentStrategy(method models.PaymentMethod) (PaymentStrategy, error) {
	switch method {
	case models.PaymentMethodUPI:
		return &UPIPayment{}, nil
	case models.PaymentMethodCreditCard:
		return &CreditCardPayment{}, nil
	case models.PaymentMethodCOD:
		return &CashOnDeliveryPayment{}, nil
	default:
		return nil, fmt.Errorf("%w: unsupported method %s", models.ErrPaymentFailed, method)
	}
}
