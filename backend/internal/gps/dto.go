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

type GpsResponse struct {
	ID        int                       `json:"id"`
	SN        string                    `json:"sn"`
	Vehicle   *vehicles.VehicleResponse `json:"vehicle"`
	CreatedAt time.Time                 `json:"created_at"`
}
