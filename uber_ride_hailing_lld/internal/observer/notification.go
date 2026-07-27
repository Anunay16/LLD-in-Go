package observer

import (
	"fmt"
	"sync"
	"uber_ride_hailing_lld/internal/models"
)

type Observer interface {
	OnTripStatusChanged(trip *models.Trip, newStatus models.TripStatus)
}

type Subject interface {
	RegisterObserver(o Observer)
	RemoveObserver(o Observer)
	NotifyObservers(trip *models.Trip, newStatus models.TripStatus)
}

type NotificationPublisher struct {
	observers []Observer
	mu        sync.RWMutex
}

func NewNotificationPublisher() *NotificationPublisher {
	return &NotificationPublisher{
		observers: make([]Observer, 0),
	}
}

func (p *NotificationPublisher) RegisterObserver(o Observer) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.observers = append(p.observers, o)
}

func (p *NotificationPublisher) RemoveObserver(o Observer) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for i, obs := range p.observers {
		if obs == o {
			p.observers = append(p.observers[:i], p.observers[i+1:]...)
			break
		}
	}
}

func (p *NotificationPublisher) NotifyObservers(trip *models.Trip, newStatus models.TripStatus) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, obs := range p.observers {
		obs.OnTripStatusChanged(trip, newStatus)
	}
}

// ConsoleNotificationObserver logs status changes to stdout
type ConsoleNotificationObserver struct {
	Name string
}

func NewConsoleNotificationObserver(name string) *ConsoleNotificationObserver {
	return &ConsoleNotificationObserver{Name: name}
}

func (o *ConsoleNotificationObserver) OnTripStatusChanged(trip *models.Trip, newStatus models.TripStatus) {
	driverName := "Unassigned"
	if trip.Driver != nil {
		driverName = trip.Driver.Name
	}
	fmt.Printf("🔔 [%s NOTIFICATION] Trip %s status changed to -> %s (Rider: %s, Driver: %s)\n",
		o.Name, trip.ID, newStatus, trip.Rider.Name, driverName)
}
