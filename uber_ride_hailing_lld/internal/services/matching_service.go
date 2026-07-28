package services

import (
	"uber_ride_hailing_lld/internal/models"
	"uber_ride_hailing_lld/internal/strategy"
)

type MatchingService interface {
	SetStrategy(newStrategy strategy.MatchingStrategy)
	FindDriver(pickup models.Location, vehicleType models.VehicleType, availableDrivers []*models.Driver) *models.Driver
}

type matchingService struct {
	strategy strategy.MatchingStrategy
}

func NewMatchingService(s strategy.MatchingStrategy) MatchingService {
	return &matchingService{strategy: s}
}

func (s *matchingService) SetStrategy(newStrategy strategy.MatchingStrategy) {
	s.strategy = newStrategy
}

func (s *matchingService) FindDriver(pickup models.Location, vehicleType models.VehicleType, availableDrivers []*models.Driver) *models.Driver {
	return s.strategy.FindDriver(pickup, vehicleType, availableDrivers)
}
