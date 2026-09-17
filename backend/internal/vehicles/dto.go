// Package vehicles
package vehicles

import "time"

type CreateVehicleRequest struct {
	Body *struct {
		PlateNumber string      `json:"plate_number"  validate:"required"`
		VehicleType VehicleType `json:"vehicle_type" validate:"required"`
		DeviceSN    string      `json:"device_sn,omitempty"`
	}
}

type Coordinate struct {
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	UpdatedAt time.Time `json:"updated_at"`
}

type VehicleResponse struct {
	ID             int         `json:"id"`
	PlateNumber    string      `json:"plate_number"`
	VehicleType    VehicleType `json:"vehicle_type"`
	DeviceSN       string      `json:"device_sn"`
	LastCoordinate *Coordinate `json:"last_coordinate"`
	CreatedAt      time.Time   `json:"created_at"`
}

type VehicleHistoryResponse struct {
	VehicleID   int           `json:"vehicle_id"`
	PlateNumber string        `json:"plate_number"`
	Coordinates *[]Coordinate `json:"coordinates"`
}

type DeleteVehicleParams struct {
	ID int `path:"id" doc:"Unique identifier for the vehicle"`
}

type GetVehicleParams struct {
	ID int `path:"id" doc:"Unique identifier for the vehicle"`
}

type GetVehicleHistoryParams struct {
	ID int `path:"id" doc:"Unique identifier for the vehicle"`
}
