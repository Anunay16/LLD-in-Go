package tests

import (
	"sync"
	"testing"

	"uber_ride_hailing_lld/internal/manager"
	"uber_ride_hailing_lld/internal/models"
	"uber_ride_hailing_lld/internal/strategy"
)

func TestRideBookingAndLifecycle(t *testing.T) {
	app := manager.NewRideHailingManager()

	r1 := app.RegisterRider("R1", "Alice", "+1-111", models.NewLocation(12.97, 77.59))
	v1 := models.NewVehicle("V1", "KA01A1234", "Civic", models.VehicleTypeSedan)
	d1 := app.RegisterDriver("D1", "Bob", "+1-222", v1, models.NewLocation(12.971, 77.591))

	pickup := models.NewLocation(12.97, 77.59)
	drop := models.NewLocation(12.98, 77.60)

	trip, err := app.BookRide("T1", r1.ID, pickup, drop, models.VehicleTypeSedan)
	if err != nil {
		t.Fatalf("Expected booking to succeed, got error: %v", err)
	}

	if trip.Driver.ID != d1.ID {
		t.Errorf("Expected driver %s, got %s", d1.ID, trip.Driver.ID)
	}

	if trip.GetStatus() != models.TripStatusAccepted {
		t.Errorf("Expected status ACCEPTED, got %s", trip.GetStatus())
	}

	if d1.GetStatus() != models.DriverStatusOnTrip {
		t.Errorf("Expected driver status ON_TRIP, got %s", d1.GetStatus())
	}

	// Start Ride
	err = app.StartRide("T1")
	if err != nil {
		t.Fatalf("Expected start ride to succeed: %v", err)
	}

	if trip.GetStatus() != models.TripStatusInProgress {
		t.Errorf("Expected status IN_PROGRESS, got %s", trip.GetStatus())
	}

	// Complete Ride
	cashPay := strategy.NewCashPaymentStrategy()
	success, msg, err := app.CompleteRide("T1", cashPay)
	if err != nil || !success {
		t.Fatalf("Expected complete ride to succeed: %v, msg: %s", err, msg)
	}

	if trip.GetStatus() != models.TripStatusCompleted {
		t.Errorf("Expected status COMPLETED, got %s", trip.GetStatus())
	}

	if d1.GetStatus() != models.DriverStatusAvailable {
		t.Errorf("Expected driver status to reset to AVAILABLE, got %s", d1.GetStatus())
	}
}

func TestConcurrentDriverMatching(t *testing.T) {
	app := manager.NewRideHailingManager()

	r1 := app.RegisterRider("R1", "Alice", "+1-111", models.NewLocation(12.97, 77.59))
	r2 := app.RegisterRider("R2", "Charlie", "+1-222", models.NewLocation(12.97, 77.59))

	// Single available driver
	v1 := models.NewVehicle("V1", "KA01A1234", "Civic", models.VehicleTypeSedan)
	_ = app.RegisterDriver("D1", "Bob", "+1-333", v1, models.NewLocation(12.971, 77.591))

	pickup := models.NewLocation(12.97, 77.59)
	drop := models.NewLocation(12.98, 77.60)

	var wg sync.WaitGroup
	wg.Add(2)

	successCount := 0
	var mu sync.Mutex

	go func() {
		defer wg.Done()
		_, err := app.BookRide("T-CONC-1", r1.ID, pickup, drop, models.VehicleTypeSedan)
		if err == nil {
			mu.Lock()
			successCount++
			mu.Unlock()
		}
	}()

	go func() {
		defer wg.Done()
		_, err := app.BookRide("T-CONC-2", r2.ID, pickup, drop, models.VehicleTypeSedan)
		if err == nil {
			mu.Lock()
			successCount++
			mu.Unlock()
		}
	}()

	wg.Wait()

	// Exactly 1 booking must succeed because only 1 driver is available!
	if successCount != 1 {
		t.Errorf("Expected exactly 1 successful booking due to race prevention, got %d", successCount)
	}
}
