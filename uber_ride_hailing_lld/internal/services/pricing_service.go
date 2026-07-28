package services

import (
	"uber_ride_hailing_lld/internal/models"
	"uber_ride_hailing_lld/internal/strategy"
)

type PricingService interface {
	SetStrategy(newStrategy strategy.PricingStrategy)
	CalculateFare(distanceKm float64, durationMinutes float64, vehicleType models.VehicleType) float64
}

type pricingService struct {
	strategy strategy.PricingStrategy
}

func NewPricingService(s strategy.PricingStrategy) PricingService {
	return &pricingService{strategy: s}
}

func (s *pricingService) SetStrategy(newStrategy strategy.PricingStrategy) {
	s.strategy = newStrategy
}

func (s *pricingService) CalculateFare(distanceKm float64, durationMinutes float64, vehicleType models.VehicleType) float64 {
	return s.strategy.CalculateFare(distanceKm, durationMinutes, vehicleType)
}
