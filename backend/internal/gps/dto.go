// Package gps
package gps

import (
	"time"

	"github.com/ize-302/beacon/backend/internal/vehicles"
)

type CreateGpsRequest struct {
	SN        string `json:"sn" validate:"required"`
	VehicleID int    `json:"vehicle_id" validate:"required"`
}

type Coordinate struct {
	Latitude  float64   `json:"latitude" validate:"required"`
	Longitude float64   `json:"longitude" validate:"required"`
	UpdatedAt time.Time `json:"updated_at" validate:"required"`
}

type GpsResponse struct {
	ID             int                       `json:"id" validate:"required"`
	SN             string                    `json:"sn" validate:"required"`
	Vehicle        *vehicles.VehicleResponse `json:"vehicle" validate:"required"`
	LastCoordinate *Coordinate               `json:"last_coordinate" validate:"required"`
	CreatedAt      time.Time                 `json:"created_at" validate:"required"`
}

type GpsHistoryResponse struct {
	GpsID       int           `json:"gps_id" validate:"required"`
	GpsSN       string        `json:"gps_sn" validate:"required"`
	Coordinates *[]Coordinate `json:"coordinates" validate:"required"`
}
