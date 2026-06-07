// Package vehicles
package vehicles

import "time"

type CreateVehicleRequest struct {
	PlateNumber string `json:"plate_number"`
}

type VehicleResponse struct {
	ID          int       `json:"id"`
	PlateNumber string    `json:"plate_number"`
	CreatedAt   time.Time `json:"created_at"`
}
