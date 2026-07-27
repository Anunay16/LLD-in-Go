package models

type DriverStatus string

const (
	DriverStatusOffline   DriverStatus = "OFFLINE"
	DriverStatusAvailable DriverStatus = "AVAILABLE"
	DriverStatusOnTrip    DriverStatus = "ON_TRIP"
)

type TripStatus string

const (
	TripStatusRequested  TripStatus = "REQUESTED"
	TripStatusAccepted   TripStatus = "ACCEPTED"
	TripStatusInProgress TripStatus = "IN_PROGRESS"
	TripStatusCompleted  TripStatus = "COMPLETED"
	TripStatusCancelled  TripStatus = "CANCELLED"
)

type VehicleType string

const (
	VehicleTypeBike  VehicleType = "BIKE"
	VehicleTypeAuto  VehicleType = "AUTO"
	VehicleTypeSedan VehicleType = "SEDAN"
	VehicleTypeSUV   VehicleType = "SUV"
)

type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "PENDING"
	PaymentStatusCompleted PaymentStatus = "COMPLETED"
	PaymentStatusFailed    PaymentStatus = "FAILED"
)

type PaymentMode string

const (
	PaymentModeCash       PaymentMode = "CASH"
	PaymentModeCreditCard PaymentMode = "CREDIT_CARD"
	PaymentModeWallet     PaymentMode = "WALLET"
)
