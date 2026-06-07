// Package locations
package locations

import (
	"time"

	"github.com/ize-302/beacon/backend/internal/vehicles"
)

type CreateLocation struct {
	VehicleID int     `json:"vehicle_id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type LocationResponse struct {
	ID        int                       `json:"id"`
	Vehicle   *vehicles.VehicleResponse `json:"vehicle"`
	Latitude  float64                   `json:"latitude"`
	Longitude float64                   `json:"longitude"`
	CreatedAt time.Time                 `json:"created_at"`
}
