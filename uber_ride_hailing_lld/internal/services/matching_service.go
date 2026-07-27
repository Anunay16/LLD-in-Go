package services

import (
	"uber_ride_hailing_lld/internal/models"
	"uber_ride_hailing_lld/internal/strategy"
)

type MatchingService struct {
	strategy strategy.MatchingStrategy
}

func NewMatchingService(s strategy.MatchingStrategy) *MatchingService {
	return &MatchingService{strategy: s}
}

func (s *MatchingService) SetStrategy(newStrategy strategy.MatchingStrategy) {
	s.strategy = newStrategy
}

func (s *MatchingService) FindDriver(pickup models.Location, vehicleType models.VehicleType, availableDrivers []*models.Driver) *models.Driver {
	return s.strategy.FindDriver(pickup, vehicleType, availableDrivers)
}
