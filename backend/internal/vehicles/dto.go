// Package vehicles
package vehicles

import "time"

type CreateVehicleRequest struct {
	PlateNumber string      `json:"plate_number"  validate:"required"`
	VehicleType VehicleType `json:"vehicle_type" validate:"required"`
}

type VehicleResponse struct {
	ID          int         `json:"id" validate:"required"`
	PlateNumber string      `json:"plate_number" validate:"required"`
	CreatedAt   time.Time   `json:"created_at" validate:"required"`
	VehicleType VehicleType `json:"vehicle_type" validate:"required"`
}
