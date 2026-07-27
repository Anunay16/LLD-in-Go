package strategy

import "uber_ride_hailing_lld/internal/models"

type PricingStrategy interface {
	CalculateFare(distance float64, durationMinutes float64, vehicleType models.VehicleType) float64
}

type StandardPricingStrategy struct {
	BaseRates map[models.VehicleType]float64
	PerKmRates map[models.VehicleType]float64
	PerMinRates map[models.VehicleType]float64
}

func NewStandardPricingStrategy() *StandardPricingStrategy {
	return &StandardPricingStrategy{
		BaseRates: map[models.VehicleType]float64{
			models.VehicleTypeBike:  15.0,
			models.VehicleTypeAuto:  25.0,
			models.VehicleTypeSedan: 50.0,
			models.VehicleTypeSUV:   80.0,
		},
		PerKmRates: map[models.VehicleType]float64{
			models.VehicleTypeBike:  8.0,
			models.VehicleTypeAuto:  12.0,
			models.VehicleTypeSedan: 18.0,
			models.VehicleTypeSUV:   25.0,
		},
		PerMinRates: map[models.VehicleType]float64{
			models.VehicleTypeBike:  1.0,
			models.VehicleTypeAuto:  1.5,
			models.VehicleTypeSedan: 2.0,
			models.VehicleTypeSUV:   3.0,
		},
	}
}

func (s *StandardPricingStrategy) CalculateFare(distance float64, durationMinutes float64, vehicleType models.VehicleType) float64 {
	base := s.BaseRates[vehicleType]
	perKm := s.PerKmRates[vehicleType]
	perMin := s.PerMinRates[vehicleType]

	return base + (distance * perKm) + (durationMinutes * perMin)
}

type SurgePricingStrategy struct {
	BaseStrategy    PricingStrategy
	SurgeMultiplier float64
}

func NewSurgePricingStrategy(baseStrategy PricingStrategy, multiplier float64) *SurgePricingStrategy {
	return &SurgePricingStrategy{
		BaseStrategy:    baseStrategy,
		SurgeMultiplier: multiplier,
	}
}

func (s *SurgePricingStrategy) CalculateFare(distance float64, durationMinutes float64, vehicleType models.VehicleType) float64 {
	baseFare := s.BaseStrategy.CalculateFare(distance, durationMinutes, vehicleType)
	return baseFare * s.SurgeMultiplier
}
