package booking

import (
	"errors"
	"time"
)

type SeatStatus string

const (
	SeatStatusAvailable SeatStatus = "AVAILABLE"
	SeatStatusHeld      SeatStatus = "HELD"
	SeatStatusBooked   SeatStatus = "BOOKED"
)

type HoldStatus string

const (
	HoldStatusActive    HoldStatus = "ACTIVE"
	HoldStatusConfirmed HoldStatus = "CONFIRMED"
	HoldStatusExpired   HoldStatus = "EXPIRED"
	HoldStatusCancelled HoldStatus = "CANCELLED"
)

var (
	ErrShowNotFound      = errors.New("show not found")
	ErrSeatNotFound      = errors.New("seat not found")
	ErrSeatAlreadyBooked = errors.New("seat already booked")
	ErrSeatHeldByOther   = errors.New("seat currently held by another user")
	ErrHoldNotFound      = errors.New("hold transaction not found")
	ErrHoldExpired       = errors.New("hold transaction has expired")
	ErrHoldNotActive     = errors.New("hold transaction is no longer active")
	ErrHoldUserMismatch  = errors.New("hold transaction belongs to another user")
)

// Seat represents an individual seat in a venue/show.
type Seat struct {
	ID        string     `json:"id"`
	Number    string     `json:"number"`
	Status    SeatStatus `json:"status"`
	HeldBy    string     `json:"held_by,omitempty"`
	HoldID    string     `json:"hold_id,omitempty"`
	HeldUntil time.Time  `json:"held_until,omitempty"`
}

// IsAvailable checks if seat is available or if its hold duration has expired.
// If the hold has expired, it lazily releases the seat by resetting its status and metadata.
func (s *Seat) IsAvailable(now time.Time) bool {
	if s.Status == SeatStatusAvailable {
		return true
	}
	if s.Status == SeatStatusHeld && now.After(s.HeldUntil) {
		s.Status = SeatStatusAvailable
		s.HeldBy = ""
		s.HoldID = ""
		s.HeldUntil = time.Time{}
		return true
	}
	return false
}

// Hold represents a temporary seat reservation prior to payment completion.
type Hold struct {
	ID        string     `json:"id"`
	ShowID    string     `json:"show_id"`
	UserID    string     `json:"user_id"`
	SeatIDs   []string   `json:"seat_ids"`
	ExpiresAt time.Time  `json:"expires_at"`
	Status    HoldStatus `json:"status"`
}

// Show represents an event/movie screening containing seats.
type Show struct {
	ID    string           `json:"id"`
	Title string           `json:"title"`
	Seats map[string]*Seat `json:"seats"` // seatID -> Seat
}
