// Package vehicles
package vehicles

import "time"

type VehicleType string

const (
	Car   VehicleType = "car"
	Bus   VehicleType = "bus"
	Truck VehicleType = "truck"
	Van   VehicleType = "van"
)

type Vehicle struct {
	ID          int
	PlateNumber string
	CreatedAt   time.Time
	VehicleType VehicleType
}
