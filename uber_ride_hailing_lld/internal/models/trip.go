package models

import (
	"sync"
	"time"
)

type Trip struct {
	ID            string
	Rider         *Rider
	Driver        *Driver
	Pickup        Location
	Dropoff       Location
	RequestedType VehicleType
	Status        TripStatus
	Fare          float64
	StartTime     time.Time
	EndTime       time.Time
	mu            sync.RWMutex
}

func NewTrip(id string, rider *Rider, pickup, dropoff Location, vType VehicleType) *Trip {
	return &Trip{
		ID:            id,
		Rider:         rider,
		Pickup:        pickup,
		Dropoff:       dropoff,
		RequestedType: vType,
		Status:        TripStatusRequested,
	}
}

func (t *Trip) GetStatus() TripStatus {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Status
}

func (t *Trip) SetStatus(status TripStatus) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Status = status
	if status == TripStatusInProgress {
		t.StartTime = time.Now()
	} else if status == TripStatusCompleted || status == TripStatusCancelled {
		t.EndTime = time.Now()
	}
}

func (t *Trip) AssignDriver(driver *Driver) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Driver = driver
	t.Status = TripStatusAccepted
}

func (t *Trip) SetFare(fare float64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Fare = fare
}
