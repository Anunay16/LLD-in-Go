package main

import (
	"fmt"

	"uber_ride_hailing_lld/internal/manager"
	"uber_ride_hailing_lld/internal/models"
	"uber_ride_hailing_lld/internal/strategy"
)

func main() {
	fmt.Println("==========================================================")
	fmt.Println("       🚗 UBER RIDE HAILING SYSTEM - LLD SIMULATION 🚗    ")
	fmt.Println("==========================================================")
	fmt.Println()

	// 1. Initialize System Facade Manager
	app := manager.NewRideHailingManager()

	// 2. Register Riders
	r1 := app.RegisterRider("R001", "Alice Smith", "+1-555-0101", models.NewLocation(12.9716, 77.5946)) // MG Road, Bengaluru
	r2 := app.RegisterRider("R002", "Bob Jones", "+1-555-0102", models.NewLocation(12.9352, 77.6245))   // Koramangala

	fmt.Printf("Registered Riders:\n - %s (Rating: %.1f)\n - %s (Rating: %.1f)\n\n", r1.Name, r1.Rating, r2.Name, r2.Rating)

	// 3. Register Vehicles & Drivers
	v1 := models.NewVehicle("V001", "KA-01-AB-1234", "Honda City", models.VehicleTypeSedan)
	v2 := models.NewVehicle("V002", "KA-05-CD-5678", "Toyota Fortuner", models.VehicleTypeSUV)
	v3 := models.NewVehicle("V003", "KA-03-EF-9012", "Bajaj RE Auto", models.VehicleTypeAuto)
	v4 := models.NewVehicle("V004", "KA-02-GH-3456", "Hyundai Verna", models.VehicleTypeSedan)

	d1 := app.RegisterDriver("D001", "Charlie", "+1-555-0201", v1, models.NewLocation(12.9720, 77.5950)) // Close to Alice
	d1.Rating = 4.8
	d2 := app.RegisterDriver("D002", "David", "+1-555-0202", v2, models.NewLocation(12.9360, 77.6250))   // SUV
	d2.Rating = 4.9
	d3 := app.RegisterDriver("D003", "Eve", "+1-555-0203", v3, models.NewLocation(12.9730, 77.5960))     // Auto
	d3.Rating = 4.5
	d4 := app.RegisterDriver("D004", "Frank", "+1-555-0204", v4, models.NewLocation(12.9750, 77.5990))   // Sedan (further away)
	d4.Rating = 5.0

	fmt.Println("Registered & Available Drivers:")
	fmt.Printf(" - %s (%s, Rating: %.1f) at %s\n", d1.Name, d1.Vehicle.Model, d1.Rating, d1.GetLocation())
	fmt.Printf(" - %s (%s, Rating: %.1f) at %s\n", d2.Name, d2.Vehicle.Model, d2.Rating, d2.GetLocation())
	fmt.Printf(" - %s (%s, Rating: %.1f) at %s\n", d3.Name, d3.Vehicle.Model, d3.Rating, d3.GetLocation())
	fmt.Printf(" - %s (%s, Rating: %.1f) at %s\n\n", d4.Name, d4.Vehicle.Model, d4.Rating, d4.GetLocation())

	// ------------------------------------------------------------------
	// SCENARIO 1: Standard Ride Request (Nearest Driver Strategy)
	// ------------------------------------------------------------------
	fmt.Println("----------------------------------------------------------")
	fmt.Println("SCENARIO 1: Alice Books a Sedan (Nearest Driver Strategy)")
	fmt.Println("----------------------------------------------------------")

	pickup1 := models.NewLocation(12.9716, 77.5946)
	dropoff1 := models.NewLocation(12.9279, 77.6271) // HSR Layout

	trip1, err := app.BookRide("TRIP-1001", r1.ID, pickup1, dropoff1, models.VehicleTypeSedan)
	if err != nil {
		fmt.Printf("❌ Failed to book ride: %v\n", err)
	} else {
		fmt.Printf("✅ Ride Booked! Driver Assigned: %s (%s, Plate: %s)\n",
			trip1.Driver.Name, trip1.Driver.Vehicle.Model, trip1.Driver.Vehicle.PlateNumber)
		fmt.Printf("💰 Calculated Fare: ₹%.2f\n", trip1.Fare)

		// Start Ride
		fmt.Println("▶ Starting Trip...")
		_ = app.StartRide(trip1.ID)

		// Complete Ride & Pay using Credit Card
		fmt.Println("🏁 Completing Trip...")
		ccPayment := strategy.NewCreditCardPaymentStrategy("4111-2222-3333-4444")
		success, msg, _ := app.CompleteRide(trip1.ID, ccPayment)
		if success {
			fmt.Printf("💳 Payment Success: %s\n", msg)
		}
	}
	fmt.Println()

	// ------------------------------------------------------------------
	// SCENARIO 2: Surge Pricing + Highest Rated Driver Strategy
	// ------------------------------------------------------------------
	fmt.Println("----------------------------------------------------------")
	fmt.Println("SCENARIO 2: Bob Books a Ride during Peak Demand (Surge Pricing 1.5x)")
	fmt.Println("----------------------------------------------------------")

	// Switch Strategy dynamically
	surgePricing := strategy.NewSurgePricingStrategy(strategy.NewStandardPricingStrategy(), 1.5)
	app.SetPricingStrategy(surgePricing)
	app.SetMatchingStrategy(strategy.NewHighestRatedDriverStrategy())

	pickup2 := models.NewLocation(12.9352, 77.6245)
	dropoff2 := models.NewLocation(12.9716, 77.5946)

	trip2, err := app.BookRide("TRIP-1002", r2.ID, pickup2, dropoff2, models.VehicleTypeSUV)
	if err != nil {
		fmt.Printf("❌ Failed to book ride: %v\n", err)
	} else {
		fmt.Printf("✅ Ride Booked! Driver Assigned: %s (%s, Rating: %.1f)\n",
			trip2.Driver.Name, trip2.Driver.Vehicle.Model, trip2.Driver.Rating)
		fmt.Printf("💰 Calculated Surge Fare (1.5x): ₹%.2f\n", trip2.Fare)

		_ = app.StartRide(trip2.ID)

		// Pay with Wallet
		walletPayment := strategy.NewWalletPaymentStrategy("WAL-9876", 500.0)
		success, msg, _ := app.CompleteRide(trip2.ID, walletPayment)
		if success {
			fmt.Printf("👛 Payment Success: %s\n", msg)
		}
	}
	fmt.Println()

	// ------------------------------------------------------------------
	// SCENARIO 3: Error Handling - No Available Driver for Bike
	// ------------------------------------------------------------------
	fmt.Println("----------------------------------------------------------")
	fmt.Println("SCENARIO 3: Requesting a Bike when none is registered")
	fmt.Println("----------------------------------------------------------")

	_, err = app.BookRide("TRIP-1003", r1.ID, pickup1, dropoff1, models.VehicleTypeBike)
	if err != nil {
		fmt.Printf("⚠️ Expected Error Captured: %v\n", err)
	}

	fmt.Println()
	fmt.Println("==========================================================")
	fmt.Println("           ✅ LLD SIMULATION COMPLETED SUCCESSFULLY       ")
	fmt.Println("==========================================================")
}
