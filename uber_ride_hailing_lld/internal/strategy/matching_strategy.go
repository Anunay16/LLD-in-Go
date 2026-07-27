package strategy

import (
	"math"
	"uber_ride_hailing_lld/internal/models"
)

type MatchingStrategy interface {
	FindDriver(pickup models.Location, vehicleType models.VehicleType, availableDrivers []*models.Driver) *models.Driver
}

// NearestDriverStrategy matches the available driver closest to the pickup location matching the vehicle type
type NearestDriverStrategy struct{}

func NewNearestDriverStrategy() *NearestDriverStrategy {
	return &NearestDriverStrategy{}
}

func (s *NearestDriverStrategy) FindDriver(pickup models.Location, vehicleType models.VehicleType, availableDrivers []*models.Driver) *models.Driver {
	var bestDriver *models.Driver
	minDistance := math.MaxFloat64

	for _, driver := range availableDrivers {
		if !driver.IsAvailable() {
			continue
		}
		if driver.Vehicle == nil || driver.Vehicle.Type != vehicleType {
			continue
		}

		dist := pickup.DistanceTo(driver.GetLocation())
		if dist < minDistance {
			minDistance = dist
			bestDriver = driver
		}
	}

	return bestDriver
}

// HighestRatedDriverStrategy matches the available driver with the highest rating matching the vehicle type
type HighestRatedDriverStrategy struct{}

func NewHighestRatedDriverStrategy() *HighestRatedDriverStrategy {
	return &HighestRatedDriverStrategy{}
}

func (s *HighestRatedDriverStrategy) FindDriver(pickup models.Location, vehicleType models.VehicleType, availableDrivers []*models.Driver) *models.Driver {
	var bestDriver *models.Driver
	highestRating := -1.0

	for _, driver := range availableDrivers {
		if !driver.IsAvailable() {
			continue
		}
		if driver.Vehicle == nil || driver.Vehicle.Type != vehicleType {
			continue
		}

		if driver.Rating > highestRating {
			highestRating = driver.Rating
			bestDriver = driver
		}
	}

	return bestDriver
}
