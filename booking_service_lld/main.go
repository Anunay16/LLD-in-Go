package main

import (
	"context"
	"fmt"
	"time"

	"booking_system/internal/booking"
	"booking_system/internal/lock"
)

func main() {
	ctx := context.Background()

	// Initialize In-Memory Lock Provider & Repository
	lockProvider := lock.NewInMemoryLockProvider(10 * time.Millisecond)
	repo := booking.NewInMemoryRepository()
	bookingService := booking.NewBookingService(lockProvider, repo, 3*time.Second, 5*time.Second)

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
	_ = repo.SaveShow(ctx, show)

	fmt.Println("==================================================")
	fmt.Println(" Movie Ticket Booking Service - Low Level Design ")
	fmt.Println("==================================================")
	fmt.Printf("Show ID: %s | Title: %s\n", show.ID, show.Title)
	printSeats(repo, showID)

	fmt.Println("\n--- Quick Demo ---")
	hold, err := bookingService.HoldSeats(ctx, showID, "Alice", []string{"A1", "A2"})
	if err != nil {
		fmt.Println("Hold failed:", err)
		return
	}
	fmt.Printf("✅ Alice held seats [A1, A2]. Hold ID: %s (Expires: %s)\n", hold.ID[:8], hold.ExpiresAt.Format("15:04:05"))

	if err := bookingService.ConfirmBooking(ctx, showID, "Alice", hold.ID); err != nil {
		fmt.Println("Confirm failed:", err)
		return
	}
	fmt.Println("🎉 Alice confirmed booking!")

	fmt.Println("\n=== Final Show Seats State ===")
	printSeats(repo, showID)

	fmt.Println("\n💡 Run 'go test -v ./...' to execute all test scenarios (High Concurrency, Expiration, Cancellation, Edge Cases).")
}

func printSeats(repo booking.Repository, showID string) {
	ctx := context.Background()
	show, err := repo.GetShow(ctx, showID)
	if err != nil {
		fmt.Println("Error reading show:", err)
		return
	}

	for _, seatID := range []string{"A1", "A2", "A3", "B1", "B2"} {
		seat := show.Seats[seatID]
		fmt.Printf("Seat %s: Status=%-9s HeldBy=%-7s HoldID=%s\n",
			seat.ID, seat.Status, seat.HeldBy, truncateID(seat.HoldID))
	}
}

func truncateID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}
