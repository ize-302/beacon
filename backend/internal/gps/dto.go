// Package gps
package gps

import (
	"time"

	"github.com/ize-302/beacon/backend/internal/vehicles"
)

type CreateGpsRequest struct {
	Body *struct {
		SN        string `json:"sn" validate:"required"`
		VehicleID int    `json:"vehicle_id" validate:"required"`
	}
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

type GpsHistoryResponse struct {
	GpsID       int          `json:"gps_id"`
	GpsSN       string       `json:"gps_sn"`
	Coordinates *[]Coordinate `json:"coordinates"`
}

type GetGpsParams struct {
	ID int `path:"id" doc:"Unique identifier for the GPS device"`
}

type DeleteGpsParams struct {
	ID int `path:"id" doc:"Unique identifier for the GPS device"`
}

type GetGpsHistoryParams struct {
	ID int `path:"id" doc:"Unique identifier for the GPS device"`
}
