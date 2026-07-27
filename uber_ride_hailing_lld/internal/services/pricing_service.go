package services

import (
	"uber_ride_hailing_lld/internal/models"
	"uber_ride_hailing_lld/internal/strategy"
)

type PricingService struct {
	strategy strategy.PricingStrategy
}

func NewPricingService(s strategy.PricingStrategy) *PricingService {
	return &PricingService{strategy: s}
}

func (s *PricingService) SetStrategy(newStrategy strategy.PricingStrategy) {
	s.strategy = newStrategy
}

func (s *PricingService) CalculateFare(distanceKm float64, durationMinutes float64, vehicleType models.VehicleType) float64 {
	return s.strategy.CalculateFare(distanceKm, durationMinutes, vehicleType)
}
