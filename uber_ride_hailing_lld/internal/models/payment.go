package models

import "time"

type Payment struct {
	ID        string
	TripID    string
	Amount    float64
	Mode      PaymentMode
	Status    PaymentStatus
	Timestamp time.Time
}

func NewPayment(id, tripID string, amount float64, mode PaymentMode) *Payment {
	return &Payment{
		ID:        id,
		TripID:    tripID,
		Amount:    amount,
		Mode:      mode,
		Status:    PaymentStatusPending,
		Timestamp: time.Now(),
	}
}
