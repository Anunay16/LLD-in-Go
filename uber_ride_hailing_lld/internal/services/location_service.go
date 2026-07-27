package services

import "uber_ride_hailing_lld/internal/models"

type LocationService struct{}

func NewLocationService() *LocationService {
	return &LocationService{}
}

func (s *LocationService) CalculateDistance(loc1, loc2 models.Location) float64 {
	return loc1.DistanceTo(loc2)
}

// EstimateDurationInMinutes assumes an average speed of 30 km/h in city traffic
func (s *LocationService) EstimateDurationInMinutes(distanceKm float64) float64 {
	if distanceKm <= 0 {
		return 1.0
	}
	const avgSpeedKmPerHour = 30.0
	hours := distanceKm / avgSpeedKmPerHour
	return hours * 60.0
}
