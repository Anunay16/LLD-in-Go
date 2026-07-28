package models

import "sync"

type User struct {
	ID              string
	Name            string
	Phone           string
	Rating          float64
	CurrentLocation Location
}

type Rider struct {
	User
}

func NewRider(id, name, phone string, initialLoc Location) *Rider {
	return &Rider{
		User: User{
			ID:              id,
			Name:            name,
			Phone:           phone,
			Rating:          5.0,
			CurrentLocation: initialLoc,
		},
	}
}

type Driver struct {
	User
	Vehicle *Vehicle
	Status  DriverStatus
	mu      sync.RWMutex
}

func NewDriver(id, name, phone string, vehicle *Vehicle, initialLoc Location) *Driver {
	return &Driver{
		User: User{
			ID:              id,
			Name:            name,
			Phone:           phone,
			Rating:          5.0,
			CurrentLocation: initialLoc,
		},
		Vehicle: vehicle,
		Status:  DriverStatusOffline,
	}
}

func (d *Driver) GetStatus() DriverStatus {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.Status
}

func (d *Driver) SetStatus(status DriverStatus) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.Status = status
}

func (d *Driver) GetLocation() Location {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.CurrentLocation
}

func (d *Driver) UpdateLocation(loc Location) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.CurrentLocation = loc
}

func (d *Driver) IsAvailable() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.Status == DriverStatusAvailable
}

// TryAssignToTrip atomically checks if the driver is AVAILABLE and sets their status to ON_TRIP.
// Returns true if assignment succeeded, or false if the driver was taken concurrently.
func (d *Driver) TryAssignToTrip() bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.Status != DriverStatusAvailable {
		return false
	}
	d.Status = DriverStatusOnTrip
	return true
}
