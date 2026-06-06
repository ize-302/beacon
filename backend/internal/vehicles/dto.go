// Package vehicles
package vehicles

type CreateVehicleRequest struct {
	PlateNumber string `json:"plate_number"`
}

type VehicleResponse struct {
	ID          int    `json:"id"`
	PlateNumber string `json:"plate_number"`
}
