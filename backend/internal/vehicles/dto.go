// Package vehicles
package vehicles

import "time"

type CreateVehicleRequest struct {
	Body *struct {
		PlateNumber string      `json:"plate_number"  validate:"required"`
		VehicleType VehicleType `json:"vehicle_type" validate:"required"`
	}
}

type VehicleResponse struct {
	ID          int         `json:"id"`
	PlateNumber string      `json:"plate_number"`
	CreatedAt   time.Time   `json:"created_at"`
	VehicleType VehicleType `json:"vehicle_type"`
}

type DeleteVehicleParams struct {
	ID int `path:"id" doc:"Unique identifier for the vehicle"`
}

type GetVehicleParams struct {
	ID int `path:"id" doc:"Unique identifier for the vehicle"`
}
