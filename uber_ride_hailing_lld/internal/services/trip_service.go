package services

import (
	"fmt"
	"sync"
	"uber_ride_hailing_lld/internal/models"
	"uber_ride_hailing_lld/internal/observer"
	"uber_ride_hailing_lld/internal/strategy"
)

type TripService interface {
	CreateTrip(tripID string, riderID string, pickup, dropoff models.Location, vType models.VehicleType) (*models.Trip, error)
	StartTrip(tripID string) error
	CompleteTrip(tripID string, paymentStrategy strategy.PaymentStrategy) (bool, string, error)
	GetTrip(tripID string) (*models.Trip, error)
}

type tripService struct {
	trips           map[string]*models.Trip
	userService     UserService
	matchingService MatchingService
	pricingService  PricingService
	locationService LocationService
	publisher       observer.Subject
	mu              sync.RWMutex
}

func NewTripService(
	userService UserService,
	matchingService MatchingService,
	pricingService PricingService,
	locationService LocationService,
	publisher observer.Subject,
) TripService {
	return &tripService{
		trips:           make(map[string]*models.Trip),
		userService:     userService,
		matchingService: matchingService,
		pricingService:  pricingService,
		locationService: locationService,
		publisher:       publisher,
	}
}

func (s *tripService) CreateTrip(tripID string, riderID string, pickup, dropoff models.Location, vType models.VehicleType) (*models.Trip, error) {
	rider, err := s.userService.GetRider(riderID)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	if _, exists := s.trips[tripID]; exists {
		s.mu.Unlock()
		return nil, fmt.Errorf("trip ID %s already exists", tripID)
	}

	trip := models.NewTrip(tripID, rider, pickup, dropoff, vType)
	s.trips[tripID] = trip
	s.mu.Unlock()

	s.publisher.NotifyObservers(trip, models.TripStatusRequested)

	// Estimate fare
	dist := s.locationService.CalculateDistance(pickup, dropoff)
	duration := s.locationService.EstimateDurationInMinutes(dist)
	fare := s.pricingService.CalculateFare(dist, duration, vType)
	trip.SetFare(fare)

	// Attempt driver matching
	availableDrivers := s.userService.GetAvailableDrivers()
	driver := s.matchingService.FindDriver(pickup, vType, availableDrivers)

	if driver == nil {
		s.publisher.NotifyObservers(trip, models.TripStatusCancelled)
		return nil, fmt.Errorf("no available driver found matching vehicle type %s near pickup location", vType)
	}

	// Thread-safe lock on driver to prevent double booking
	if !s.assignDriverSafely(trip, driver) {
		return nil, fmt.Errorf("driver %s was locked by another trip", driver.ID)
	}

	return trip, nil
}

func (s *tripService) assignDriverSafely(trip *models.Trip, driver *models.Driver) bool {
	// Atomic check-and-set to prevent TOCTOU race conditions when multiple trips match the same driver
	if !driver.TryAssignToTrip() {
		return false
	}

	trip.AssignDriver(driver)
	s.publisher.NotifyObservers(trip, models.TripStatusAccepted)
	return true
}

func (s *tripService) StartTrip(tripID string) error {
	s.mu.RLock()
	trip, exists := s.trips[tripID]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("trip %s not found", tripID)
	}

	if trip.GetStatus() != models.TripStatusAccepted {
		return fmt.Errorf("cannot start trip %s from status %s", tripID, trip.GetStatus())
	}

	trip.SetStatus(models.TripStatusInProgress)
	s.publisher.NotifyObservers(trip, models.TripStatusInProgress)
	return nil
}

func (s *tripService) CompleteTrip(tripID string, paymentStrategy strategy.PaymentStrategy) (bool, string, error) {
	s.mu.RLock()
	trip, exists := s.trips[tripID]
	s.mu.RUnlock()

	if !exists {
		return false, "", fmt.Errorf("trip %s not found", tripID)
	}

	if trip.GetStatus() != models.TripStatusInProgress {
		return false, "", fmt.Errorf("cannot complete trip %s from status %s", tripID, trip.GetStatus())
	}

	// Update trip & driver status
	trip.SetStatus(models.TripStatusCompleted)
	if trip.Driver != nil {
		trip.Driver.SetStatus(models.DriverStatusAvailable)
		trip.Driver.UpdateLocation(trip.Dropoff)
	}

	s.publisher.NotifyObservers(trip, models.TripStatusCompleted)

	// Process Payment
	success, msg := paymentStrategy.ProcessPayment(tripID, trip.Fare)
	return success, msg, nil
}

func (s *tripService) GetTrip(tripID string) (*models.Trip, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	trip, exists := s.trips[tripID]
	if !exists {
		return nil, fmt.Errorf("trip %s not found", tripID)
	}
	return trip, nil
}
