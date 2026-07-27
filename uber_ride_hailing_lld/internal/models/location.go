package models

import (
	"fmt"
	"math"
)

type Location struct {
	Latitude  float64
	Longitude float64
}

func NewLocation(lat, long float64) Location {
	return Location{
		Latitude:  lat,
		Longitude: long,
	}
}

// DistanceTo calculates Euclidean distance between two locations (in kilometers approximate)
func (l Location) DistanceTo(other Location) float64 {
	latDiff := l.Latitude - other.Latitude
	longDiff := l.Longitude - other.Longitude
	// Approximate distance calculation formula for simple LLD simulation
	return math.Sqrt(latDiff*latDiff + longDiff*longDiff)
}

func (l Location) String() string {
	return fmt.Sprintf("(%.4f, %.4f)", l.Latitude, l.Longitude)
}
