package models

type Vehicle struct {
	ID          string
	PlateNumber string
	Model       string
	Type        VehicleType
}

func NewVehicle(id, plateNumber, modelName string, vType VehicleType) *Vehicle {
	return &Vehicle{
		ID:          id,
		PlateNumber: plateNumber,
		Model:       modelName,
		Type:        vType,
	}
}
