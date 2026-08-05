package main

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"booking_system/internal/booking"
	"booking_system/internal/lock"
)

// helper to create a fresh booking service & test show
func setupTestEnv(holdDuration, lockTTL time.Duration) (*booking.BookingService, booking.Repository, string) {
	lockProvider := lock.NewInMemoryLockProvider(5 * time.Millisecond)
	repo := booking.NewInMemoryRepository()
	service := booking.NewBookingService(lockProvider, repo, holdDuration, lockTTL)

	showID := "show-101"
	show := &booking.Show{
		ID:    showID,
		Title: "Avengers: Secret Wars (IMAX)",
		Seats: map[string]*booking.Seat{
			"A1": {ID: "A1", Number: "A1", Status: booking.SeatStatusAvailable},
			"A2": {ID: "A2", Number: "A2", Status: booking.SeatStatusAvailable},
			"A3": {ID: "A3", Number: "A3", Status: booking.SeatStatusAvailable},
			"B1": {ID: "B1", Number: "B1", Status: booking.SeatStatusAvailable},
			"B2": {ID: "B2", Number: "B2", Status: booking.SeatStatusAvailable},
		},
	}
	_ = repo.SaveShow(context.Background(), show)

	return service, repo, showID
}

// Scenario 1: High Concurrency Seat Booking & Deadlock Prevention
// Simulates Alice [A1, A2], Bob [A2, A3], and Charlie [A2, A1] competing concurrently for overlapping seats.
func TestScenario_HighConcurrencyBooking(t *testing.T) {
	service, repo, showID := setupTestEnv(3*time.Second, 5*time.Second)
	ctx := context.Background()

	requests := []struct {
		user  string
		seats []string
	}{
		{user: "Alice", seats: []string{"A1", "A2"}},
		{user: "Bob", seats: []string{"A2", "A3"}},
		{user: "Charlie", seats: []string{"A2", "A1"}}, // Opposite order to prove deadlock prevention!
	}

	var wg sync.WaitGroup
	type result struct {
		user string
		hold *booking.Hold
		err  error
	}
	results := make(chan result, len(requests))

	t.Log("--- Running High Concurrency Booking Scenario ---")
	for _, req := range requests {
		wg.Add(1)
		go func(u string, s []string) {
			defer wg.Done()
			hold, err := service.HoldSeats(ctx, showID, u, s)
			results <- result{user: u, hold: hold, err: err}
		}(req.user, req.seats)
	}

	wg.Wait()
	close(results)

	successCount := 0
	var winnerUser string
	var winnerHold *booking.Hold

	for res := range results {
		if res.err == nil {
			successCount++
			winnerUser = res.user
			winnerHold = res.hold
			t.Logf("✅ %s SUCCESSFULLY held seats! HoldID: %s", res.user, res.hold.ID[:8])
		} else {
			t.Logf("❌ %s failed as expected due to lock contention: %v", res.user, res.err)
		}
	}

	if successCount != 1 {
		t.Fatalf("Expected exactly 1 user to win the hold under contention, got %d", successCount)
	}

	// Verify show state in repository
	show, err := repo.GetShow(ctx, showID)
	if err != nil {
		t.Fatalf("Failed to fetch show state: %v", err)
	}

	if show.Seats["A2"].Status != booking.SeatStatusHeld || show.Seats["A2"].HeldBy != winnerUser {
		t.Errorf("Expected seat A2 to be held by winner %s, got heldBy=%s", winnerUser, show.Seats["A2"].HeldBy)
	}

	_ = winnerHold
}

// Scenario 2: Hold Seats and Confirm Booking (Happy Path)
func TestScenario_HoldAndConfirmBooking(t *testing.T) {
	service, repo, showID := setupTestEnv(3*time.Second, 5*time.Second)
	ctx := context.Background()

	t.Log("--- Running Hold and Confirm Booking Scenario ---")
	user := "Alice"
	seats := []string{"A1", "A2"}

	// 1. Hold seats
	hold, err := service.HoldSeats(ctx, showID, user, seats)
	if err != nil {
		t.Fatalf("Alice failed to hold seats: %v", err)
	}
	t.Logf("✅ Alice held seats %v with HoldID: %s", seats, hold.ID[:8])

	// 2. Confirm booking after simulated payment
	time.Sleep(50 * time.Millisecond)
	err = service.ConfirmBooking(ctx, showID, user, hold.ID)
	if err != nil {
		t.Fatalf("Alice failed to confirm booking: %v", err)
	}
	t.Logf("🎉 Alice successfully confirmed booking for seats %v!", seats)

	// 3. Verify final state in repo
	show, err := repo.GetShow(ctx, showID)
	if err != nil {
		t.Fatalf("Failed to get show: %v", err)
	}

	for _, seatID := range seats {
		seat := show.Seats[seatID]
		if seat.Status != booking.SeatStatusBooked {
			t.Errorf("Expected seat %s to be BOOKED, got %s", seatID, seat.Status)
		}
	}
}

// Scenario 3: Seat Hold Expiration and Subsequent Re-Booking
func TestScenario_HoldExpiration(t *testing.T) {
	// Short hold TTL of 60ms
	service, _, showID := setupTestEnv(60*time.Millisecond, 500*time.Millisecond)
	ctx := context.Background()

	t.Log("--- Running Seat Hold Expiration Scenario ---")

	// 1. Alice holds seat A1
	holdAlice, err := service.HoldSeats(ctx, showID, "Alice", []string{"A1"})
	if err != nil {
		t.Fatalf("Alice failed to hold seat A1: %v", err)
	}
	t.Logf("✅ Alice held seat A1 until %s", holdAlice.ExpiresAt.Format("15:04:05.000"))

	// 2. Bob immediately tries to hold A1 (should fail)
	_, err = service.HoldSeats(ctx, showID, "Bob", []string{"A1"})
	if err == nil {
		t.Fatal("Expected Bob's hold attempt to fail while Alice's hold is active")
	}
	t.Logf("❌ Bob's immediate hold attempt failed as expected: %v", err)

	// 3. Wait for Alice's hold to expire
	time.Sleep(80 * time.Millisecond)

	// 4. Bob tries to hold A1 after expiration (should succeed)
	holdBob, err := service.HoldSeats(ctx, showID, "Bob", []string{"A1"})
	if err != nil {
		t.Fatalf("Bob failed to hold seat A1 after Alice's hold expired: %v", err)
	}
	t.Logf("🎉 Bob successfully held seat A1 after Alice's hold expired! HoldID: %s", holdBob.ID[:8])
}

// Scenario 4: Early Cancellation of Seat Hold
func TestScenario_CancelHold(t *testing.T) {
	service, repo, showID := setupTestEnv(3*time.Second, 5*time.Second)
	ctx := context.Background()

	t.Log("--- Running Hold Cancellation Scenario ---")

	// 1. Bob holds seats B1, B2
	hold, err := service.HoldSeats(ctx, showID, "Bob", []string{"B1", "B2"})
	if err != nil {
		t.Fatalf("Bob failed to hold seats: %v", err)
	}
	t.Logf("✅ Bob held seats B1, B2. HoldID: %s", hold.ID[:8])

	// 2. Bob cancels hold early
	err = service.CancelHold(ctx, showID, "Bob", hold.ID)
	if err != nil {
		t.Fatalf("Bob failed to cancel hold: %v", err)
	}
	t.Log("🔄 Bob cancelled hold early.")

	// 3. Verify seats are AVAILABLE again
	show, err := repo.GetShow(ctx, showID)
	if err != nil {
		t.Fatalf("Failed to fetch show: %v", err)
	}

	for _, seatID := range []string{"B1", "B2"} {
		seat := show.Seats[seatID]
		if seat.Status != booking.SeatStatusAvailable {
			t.Errorf("Expected seat %s to be AVAILABLE, got %s", seatID, seat.Status)
		}
	}
}

// Scenario 5: Error Handling and Validation Edge Cases
func TestScenario_ValidationAndErrorHandling(t *testing.T) {
	service, _, showID := setupTestEnv(3*time.Second, 5*time.Second)
	ctx := context.Background()

	t.Log("--- Running Validation & Edge Cases Scenario ---")

	t.Run("Non-existent Show", func(t *testing.T) {
		_, err := service.HoldSeats(ctx, "invalid-show", "Alice", []string{"A1"})
		if !errors.Is(err, booking.ErrShowNotFound) {
			t.Errorf("Expected ErrShowNotFound, got %v", err)
		}
	})

	t.Run("Non-existent Seat", func(t *testing.T) {
		_, err := service.HoldSeats(ctx, showID, "Alice", []string{"Z99"})
		if !errors.Is(err, booking.ErrSeatNotFound) {
			t.Errorf("Expected ErrSeatNotFound, got %v", err)
		}
	})

	t.Run("Confirm Hold User Mismatch", func(t *testing.T) {
		hold, err := service.HoldSeats(ctx, showID, "Alice", []string{"A1"})
		if err != nil {
			t.Fatalf("Alice failed to hold seat A1: %v", err)
		}
		err = service.ConfirmBooking(ctx, showID, "Bob", hold.ID)
		if !errors.Is(err, booking.ErrHoldUserMismatch) {
			t.Errorf("Expected ErrHoldUserMismatch, got %v", err)
		}
	})

	t.Run("Hold Already Booked Seat", func(t *testing.T) {
		hold, err := service.HoldSeats(ctx, showID, "Alice", []string{"B1"})
		if err != nil {
			t.Fatalf("Alice failed to hold seat B1: %v", err)
		}
		if err := service.ConfirmBooking(ctx, showID, "Alice", hold.ID); err != nil {
			t.Fatalf("Alice failed to confirm B1: %v", err)
		}

		_, err = service.HoldSeats(ctx, showID, "Bob", []string{"B1"})
		if !errors.Is(err, booking.ErrSeatAlreadyBooked) {
			t.Errorf("Expected ErrSeatAlreadyBooked, got %v", err)
		}
	})
}
