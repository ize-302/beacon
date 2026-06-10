// Package gps
package gps

import (
	"time"

	"github.com/ize-302/beacon/backend/internal/vehicles"
)

type CreateGpsRequest struct {
	SN        string `json:"sn"`
	VehicleID int    `json:"vehicle_id"`
}

type Coordinate struct {
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	UpdatedAt time.Time `json:"updated_at"`
}

type GpsResponse struct {
	ID             int                       `json:"id"`
	SN             string                    `json:"sn"`
	Vehicle        *vehicles.VehicleResponse `json:"vehicle"`
	LastCoordinate *Coordinate               `json:"last_coordinate"`
	CreatedAt      time.Time                 `json:"created_at"`
}
