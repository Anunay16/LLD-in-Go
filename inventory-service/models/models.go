package models

import (
	"errors"
	"time"
)

var (
	ErrMedicineNotFound  = errors.New("medicine not found")
	ErrInsufficientStock = errors.New("insufficient stock available")
	ErrCartEmpty         = errors.New("cart is empty")
	ErrPaymentFailed     = errors.New("payment failed")
	ErrOrderNotFound     = errors.New("order not found")
)

// Medicine represents a pharmacy product
type Medicine struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

// CartItem represents an item in the user's cart
type CartItem struct {
	Medicine Medicine `json:"medicine"`
	Quantity int      `json:"quantity"`
}

func (ci CartItem) Subtotal() float64 {
	return float64(ci.Quantity) * ci.Medicine.Price
}

// Cart represents a user's shopping cart
type Cart struct {
	UserID string              `json:"user_id"`
	Items  map[string]CartItem `json:"items"` // Key: MedicineID
}

func NewCart(userID string) *Cart {
	return &Cart{
		UserID: userID,
		Items:  make(map[string]CartItem),
	}
}

type OrderStatus string

const (
	OrderStatusPlaced    OrderStatus = "PLACED"
	OrderStatusCancelled OrderStatus = "CANCELLED"
	OrderStatusFailed    OrderStatus = "FAILED"
)

type PaymentMethod string

const (
	PaymentMethodUPI        PaymentMethod = "UPI"
	PaymentMethodCreditCard PaymentMethod = "CREDIT_CARD"
	PaymentMethodCOD        PaymentMethod = "CASH_ON_DELIVERY"
)

type PaymentDetails struct {
	Method     PaymentMethod `json:"method"`
	CardNumber string        `json:"card_number,omitempty"`
	CVV        string        `json:"cvv,omitempty"`
	UPIID      string        `json:"upi_id,omitempty"`
}

type OrderItem struct {
	MedicineID   string  `json:"medicine_id"`
	MedicineName string  `json:"medicine_name"`
	Quantity     int     `json:"quantity"`
	UnitPrice    float64 `json:"unit_price"`
	TotalPrice   float64 `json:"total_price"`
}

type Order struct {
	ID            string        `json:"id"`
	UserID        string        `json:"user_id"`
	Items         []OrderItem   `json:"items"`
	TotalAmount   float64       `json:"total_amount"`
	PaymentMethod PaymentMethod `json:"payment_method"`
	Status        OrderStatus   `json:"status"`
	CreatedAt     time.Time     `json:"created_at"`
}
