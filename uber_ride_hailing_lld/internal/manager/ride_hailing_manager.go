package manager

import (
	"uber_ride_hailing_lld/internal/models"
	"uber_ride_hailing_lld/internal/observer"
	"uber_ride_hailing_lld/internal/services"
	"uber_ride_hailing_lld/internal/strategy"
)

type RideHailingManager struct {
	UserService     services.UserService
	MatchingService services.MatchingService
	PricingService  services.PricingService
	LocationService services.LocationService
	TripService     services.TripService
	Publisher       observer.Subject
}

func NewRideHailingManager(
	userSvc services.UserService,
	matchingSvc services.MatchingService,
	pricingSvc services.PricingService,
	locationSvc services.LocationService,
	tripSvc services.TripService,
	publisher observer.Subject,
) *RideHailingManager {
	return &RideHailingManager{
		UserService:     userSvc,
		MatchingService: matchingSvc,
		PricingService:  pricingSvc,
		LocationService: locationSvc,
		TripService:     tripSvc,
		Publisher:       publisher,
	}
}

func (m *RideHailingManager) SetMatchingStrategy(s strategy.MatchingStrategy) {
	m.MatchingService.SetStrategy(s)
}

func (m *RideHailingManager) SetPricingStrategy(s strategy.PricingStrategy) {
	m.PricingService.SetStrategy(s)
}

func (m *RideHailingManager) RegisterRider(id, name, phone string, loc models.Location) *models.Rider {
	return m.UserService.RegisterRider(id, name, phone, loc)
}

func (m *RideHailingManager) RegisterDriver(id, name, phone string, vehicle *models.Vehicle, loc models.Location) *models.Driver {
	driver := m.UserService.RegisterDriver(id, name, phone, vehicle, loc)
	driver.SetStatus(models.DriverStatusAvailable)
	return driver
}

func (m *RideHailingManager) BookRide(tripID, riderID string, pickup, dropoff models.Location, vType models.VehicleType) (*models.Trip, error) {
	return m.TripService.CreateTrip(tripID, riderID, pickup, dropoff, vType)
}

func (m *RideHailingManager) StartRide(tripID string) error {
	return m.TripService.StartTrip(tripID)
}

func (m *RideHailingManager) CompleteRide(tripID string, paymentStrategy strategy.PaymentStrategy) (bool, string, error) {
	return m.TripService.CompleteTrip(tripID, paymentStrategy)
}
