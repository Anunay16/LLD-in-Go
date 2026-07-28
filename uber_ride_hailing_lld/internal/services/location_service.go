package services

import "uber_ride_hailing_lld/internal/models"

type LocationService interface {
	CalculateDistance(loc1, loc2 models.Location) float64
	EstimateDurationInMinutes(distanceKm float64) float64
}

type locationService struct{}

func NewLocationService() LocationService {
	return &locationService{}
}

func (s *locationService) CalculateDistance(loc1, loc2 models.Location) float64 {
	return loc1.DistanceTo(loc2)
}

// EstimateDurationInMinutes assumes an average speed of 30 km/h in city traffic
func (s *locationService) EstimateDurationInMinutes(distanceKm float64) float64 {
	if distanceKm <= 0 {
		return 1.0
	}
	const avgSpeedKmPerHour = 30.0
	hours := distanceKm / avgSpeedKmPerHour
	return hours * 60.0
}
