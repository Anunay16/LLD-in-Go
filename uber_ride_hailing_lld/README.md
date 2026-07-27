# Uber Low-Level Design (LLD) in Go 🚗💨

A clean, production-ready, extensible Low-Level Design (LLD) implementation of an **Uber-like Ride Hailing System** in **Go (Golang)**. This project demonstrates core Object-Oriented Design Principles (SOLID), Software Design Patterns, Thread-Safety, and Idiomatic Go project structure using `internal/` packages.

---

## 🌟 Key Features

1. **User Management**:
   - Register **Riders** and **Drivers**.
   - Track driver availability (`OFFLINE`, `AVAILABLE`, `ON_TRIP`).
   - Track locations (`Latitude`, `Longitude`) for riders and drivers.

2. **Vehicle Management**:
   - Support multiple vehicle categories: `BIKE`, `AUTO`, `SEDAN`, `SUV`.

3. **Pluggable Driver Matching (Strategy Pattern)**:
   - Match riders with drivers based on configurable strategies:
     - `NearestDriverStrategy`: Picks the closest available driver matching vehicle type.
     - `HighestRatedDriverStrategy`: Picks the highest-rated available driver.

4. **Dynamic Fare Calculation Engine (Strategy Pattern)**:
   - Dynamic fare pricing based on distance, duration, and vehicle type:
     - `StandardPricingStrategy`: Base rate + per-KM rate + per-minute rate.
     - `SurgePricingStrategy`: Wraps standard strategy with demand multiplier (e.g., `1.5x`).

5. **Real-time Event Notifications (Observer Pattern)**:
   - Publishes trip status changes (`REQUESTED`, `ACCEPTED`, `IN_PROGRESS`, `COMPLETED`, `CANCELLED`) to subscribed notification listeners (`RIDER_APP`, `DRIVER_APP`).

6. **Pluggable Payment Gateway (Strategy Pattern)**:
   - Support multiple payment modes:
     - `CashPaymentStrategy`
     - `CreditCardPaymentStrategy`
     - `WalletPaymentStrategy` (with balance checks)

7. **Concurrency & Thread Safety**:
   - Mutex synchronization (`sync.RWMutex`) to guarantee thread-safe operations and prevent double-booking of drivers when concurrent requests arrive.

8. **Unified API Facade (Facade Pattern)**:
   - `RideHailingManager` encapsulates internal services into simple high-level API methods.

---

## 🏗 System Architecture & Class Diagram

```mermaid
classDiagram
    class Location {
        +float64 Latitude
        +float64 Longitude
        +DistanceTo(other Location) float64
    }

    class User {
        +string ID
        +string Name
        +string Phone
        +float64 Rating
        +Location CurrentLocation
    }

    class Rider {
        +User Base
    }

    class Driver {
        +User Base
        +Vehicle Vehicle
        +DriverStatus Status
        +bool IsAvailable()
    }

    class Vehicle {
        +string ID
        +string PlateNumber
        +VehicleType Type
    }

    class Trip {
        +string ID
        +Rider Rider
        +Driver Driver
        +Location Pickup
        +Location Dropoff
        +VehicleType RequestedType
        +TripStatus Status
        +float64 Fare
        +time.Time StartTime
        +time.Time EndTime
    }

    class DriverMatchingStrategy {
        <<interface>>
        +FindDriver(pickup Location, vehicleType VehicleType, drivers []*Driver) *Driver
    }

    class NearestDriverStrategy {
        +FindDriver(...) *Driver
    }

    class PricingStrategy {
        <<interface>>
        +CalculateFare(distance float64, durationMinutes float64, vType VehicleType) float64
    }

    class StandardPricingStrategy {
        +CalculateFare(...) float64
    }

    class SurgePricingStrategy {
        +CalculateFare(...) float64
    }

    class PaymentStrategy {
        <<interface>>
        +ProcessPayment(tripID string, amount float64) (bool, string)
    }

    User <|-- Rider
    User <|-- Driver
    Driver *-- Vehicle
    Trip *-- Rider
    Trip *-- Driver
    DriverMatchingStrategy <|.. NearestDriverStrategy
    PricingStrategy <|.. StandardPricingStrategy
    PricingStrategy <|.. SurgePricingStrategy
```

---

## 📁 Project Directory Layout

```
uber_ride_hailing_lld/
├── go.mod
├── README.md
├── main.go                     # Interactive simulation demo
├── internal/                   # Idiomatic Go internal package encapsulation
│   ├── models/                 # Core domain entities & enums
│   │   ├── enums.go
│   │   ├── location.go
│   │   ├── user.go
│   │   ├── vehicle.go
│   │   ├── trip.go
│   │   └── payment.go
│   ├── strategy/               # Strategy Pattern implementations
│   │   ├── matching_strategy.go
│   │   ├── pricing_strategy.go
│   │   └── payment_strategy.go
│   ├── observer/               # Observer Pattern implementation
│   │   └── notification.go
│   ├── services/               # Core business services
│   │   ├── location_service.go
│   │   ├── user_service.go
│   │   ├── pricing_service.go
│   │   ├── matching_service.go
│   │   └── trip_service.go
│   └── manager/                # Facade Manager
│       └── ride_hailing_manager.go
└── tests/
    └── trip_service_test.go    # Unit tests & race condition tests
```

---

## 🚀 Getting Started

### Prerequisites
- **Go** 1.20 or later installed on your machine.

### Running the Interactive Simulation
Execute `main.go` to observe end-to-end scenarios (rider booking, driver assignment, surge pricing, payment processing, error handling):

```bash
go run main.go
```

### Running Unit Tests & Race Condition Checks
Run standard test suite:
```bash
go test -v ./...
```

Run race detector to verify zero race conditions during concurrent driver bookings:
```bash
go test -race ./...
```

---

## 🧪 Sample Simulation Output

```text
==========================================================
       🚗 UBER RIDE HAILING SYSTEM - LLD SIMULATION 🚗    
==========================================================

Registered Riders:
 - Alice Smith (Rating: 5.0)
 - Bob Jones (Rating: 5.0)

Registered & Available Drivers:
 - Charlie (Honda City, Rating: 4.8) at (12.9720, 77.5950)
 - David (Toyota Fortuner, Rating: 4.9) at (12.9360, 77.6250)
 - Eve (Bajaj RE Auto, Rating: 4.5) at (12.9730, 77.5960)
 - Frank (Hyundai Verna, Rating: 5.0) at (12.9750, 77.5990)

----------------------------------------------------------
SCENARIO 1: Alice Books a Sedan (Nearest Driver Strategy)
----------------------------------------------------------
🔔 [RIDER_APP NOTIFICATION] Trip TRIP-1001 status changed to -> REQUESTED
🔔 [DRIVER_APP NOTIFICATION] Trip TRIP-1001 status changed to -> REQUESTED
🔔 [RIDER_APP NOTIFICATION] Trip TRIP-1001 status changed to -> ACCEPTED
🔔 [DRIVER_APP NOTIFICATION] Trip TRIP-1001 status changed to -> ACCEPTED
✅ Ride Booked! Driver Assigned: Charlie (Honda City, Plate: KA-01-AB-1234)
💰 Calculated Fare: ₹51.20
▶ Starting Trip...
🔔 [RIDER_APP NOTIFICATION] Trip TRIP-1001 status changed to -> IN_PROGRESS
🏁 Completing Trip...
🔔 [RIDER_APP NOTIFICATION] Trip TRIP-1001 status changed to -> COMPLETED
💳 Payment Success: Paid ₹51.20 via CREDIT CARD (ending 4444) for trip TRIP-1001
```

---

## 💡 Design Highlights & Trade-offs

- **Extensibility**: Adding a new pricing strategy (e.g., `DiscountPricingStrategy` or `UberPool`) or matching strategy (e.g., `ETA-based Matching`) requires zero changes to core trip service logic—simply implement the interface.
- **Encapsulation**: Using `internal/` ensures internal data models and services cannot be imported by external Go modules.
- **Thread Safety**: High concurrency handling during driver assignment prevents double-booking race conditions.
