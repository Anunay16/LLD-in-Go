package booking

import (
	"context"
	"fmt"
	"time"

	"booking_system/internal/lock"

	"github.com/google/uuid"
)

type BookingService struct {
	lockProvider lock.LockProvider
	repo         Repository
	holdDuration time.Duration
	lockTTL      time.Duration
}

func NewBookingService(lockProvider lock.LockProvider, repo Repository, holdDuration, lockTTL time.Duration) *BookingService {
	if holdDuration <= 0 {
		holdDuration = 5 * time.Minute
	}
	if lockTTL <= 0 {
		lockTTL = 10 * time.Second
	}
	return &BookingService{
		lockProvider: lockProvider,
		repo:         repo,
		holdDuration: holdDuration,
		lockTTL:      lockTTL,
	}
}

// HoldSeats attempts to temporarily reserve a set of seats for a user.
// Uses distributed/in-memory concurrency locks with canonical sorting to guarantee deadlock-freedom.
func (s *BookingService) HoldSeats(ctx context.Context, showID string, userID string, seatIDs []string) (*Hold, error) {
	if len(seatIDs) == 0 {
		return nil, fmt.Errorf("no seats specified")
	}

	// Step 1: Construct lock keys for all requested seats
	lockKeys := make([]string, len(seatIDs))
	for i, seatID := range seatIDs {
		lockKeys[i] = fmt.Sprintf("show:%s:seat:%s", showID, seatID)
	}

	// Step 2: Acquire locks concurrently for all seats.
	// TryAcquireMulti sorts keys canonically (lexicographically) before acquisition to prevent deadlocks!
	locks, err := s.lockProvider.TryAcquireMulti(ctx, lockKeys, userID, s.lockTTL)
	if err != nil {
		return nil, fmt.Errorf("failed to lock seats (concurrency contention): %w", err)
	}

	// Defer releasing concurrency locks once state mutation completes
	defer func() {
		for _, l := range locks {
			_ = l.Unlock(ctx)
		}
	}()

	// Step 3: Fetch current show state from repository
	show, err := s.repo.GetShow(ctx, showID)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	// Step 4: Validate seat availability under lock protection
	for _, seatID := range seatIDs {
		seat, exists := show.Seats[seatID]
		if !exists {
			return nil, fmt.Errorf("%w: %s", ErrSeatNotFound, seatID)
		}

		if !seat.IsAvailable(now) {
			if seat.Status == SeatStatusBooked {
				return nil, fmt.Errorf("%w: seat %s", ErrSeatAlreadyBooked, seatID)
			}
			return nil, fmt.Errorf("%w: seat %s", ErrSeatHeldByOther, seatID)
		}
	}

	// Step 5: Transition seat state to HELD
	holdID := uuid.NewString()
	expiresAt := now.Add(s.holdDuration)

	for _, seatID := range seatIDs {
		seat := show.Seats[seatID]
		seat.Status = SeatStatusHeld
		seat.HeldBy = userID
		seat.HoldID = holdID
		seat.HeldUntil = expiresAt
	}

	// Step 6: Create Hold record
	hold := &Hold{
		ID:        holdID,
		ShowID:    showID,
		UserID:    userID,
		SeatIDs:   seatIDs,
		ExpiresAt: expiresAt,
		Status:    HoldStatusActive,
	}

	// Step 7: Persist updated state atomically
	if err := s.repo.SaveShow(ctx, show); err != nil {
		return nil, fmt.Errorf("failed to save show state: %w", err)
	}

	if err := s.repo.CreateHold(ctx, hold); err != nil {
		return nil, fmt.Errorf("failed to create hold: %w", err)
	}

	return hold, nil
}

// ConfirmBooking converts an active Hold into a permanent BOOKED state upon payment.
func (s *BookingService) ConfirmBooking(ctx context.Context, showID string, userID string, holdID string) error {
	// Step 1: Retrieve Hold details
	hold, err := s.repo.GetHold(ctx, holdID)
	if err != nil {
		return err
	}

	if hold.UserID != userID {
		return ErrHoldUserMismatch
	}
	if hold.Status != HoldStatusActive {
		return ErrHoldNotActive
	}
	if time.Now().After(hold.ExpiresAt) {
		return ErrHoldExpired
	}

	// Step 2: Acquire locks for the held seats
	lockKeys := make([]string, len(hold.SeatIDs))
	for i, seatID := range hold.SeatIDs {
		lockKeys[i] = fmt.Sprintf("show:%s:seat:%s", showID, seatID)
	}

	locks, err := s.lockProvider.TryAcquireMulti(ctx, lockKeys, userID, s.lockTTL)
	if err != nil {
		return fmt.Errorf("concurrency lock error during confirmation: %w", err)
	}
	defer func() {
		for _, l := range locks {
			_ = l.Unlock(ctx)
		}
	}()

	// Step 3: Retrieve show state and confirm seat status matches hold
	show, err := s.repo.GetShow(ctx, showID)
	if err != nil {
		return err
	}

	for _, seatID := range hold.SeatIDs {
		seat, exists := show.Seats[seatID]
		if !exists || seat.HoldID != holdID || seat.HeldBy != userID {
			return fmt.Errorf("seat %s hold mismatch or expired", seatID)
		}
	}

	// Step 4: Finalize booking status
	for _, seatID := range hold.SeatIDs {
		seat := show.Seats[seatID]
		seat.Status = SeatStatusBooked
		seat.HeldBy = ""
		seat.HoldID = ""
		seat.HeldUntil = time.Time{}
	}

	hold.Status = HoldStatusConfirmed

	// Step 5: Save changes
	if err := s.repo.SaveShow(ctx, show); err != nil {
		return err
	}
	return s.repo.UpdateHold(ctx, hold)
}

// CancelHold releases a hold early (e.g. user cancels payment screen).
func (s *BookingService) CancelHold(ctx context.Context, showID string, userID string, holdID string) error {
	hold, err := s.repo.GetHold(ctx, holdID)
	if err != nil {
		return err
	}

	if hold.UserID != userID {
		return ErrHoldUserMismatch
	}

	lockKeys := make([]string, len(hold.SeatIDs))
	for i, seatID := range hold.SeatIDs {
		lockKeys[i] = fmt.Sprintf("show:%s:seat:%s", showID, seatID)
	}

	locks, err := s.lockProvider.TryAcquireMulti(ctx, lockKeys, userID, s.lockTTL)
	if err != nil {
		return fmt.Errorf("concurrency lock error during cancellation: %w", err)
	}
	defer func() {
		for _, l := range locks {
			_ = l.Unlock(ctx)
		}
	}()

	show, err := s.repo.GetShow(ctx, showID)
	if err != nil {
		return err
	}

	for _, seatID := range hold.SeatIDs {
		seat, exists := show.Seats[seatID]
		if exists && seat.HoldID == holdID {
			seat.Status = SeatStatusAvailable
			seat.HeldBy = ""
			seat.HoldID = ""
			seat.HeldUntil = time.Time{}
		}
	}

	hold.Status = HoldStatusCancelled

	if err := s.repo.SaveShow(ctx, show); err != nil {
		return err
	}
	return s.repo.UpdateHold(ctx, hold)
}
