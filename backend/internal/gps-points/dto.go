// Package gpspoints
package gpspoints

import (
	"time"
)

type CreateGpsPoint struct {
	GpsID     int     `json:"gps_id"  validate:"required"`
	Bearing   float64 `json:"bearing"  validate:"required"`
	Latitude  float64 `json:"latitude"  validate:"required"`
	Longitude float64 `json:"longitude"  validate:"required"`
	Timestamp int64   `json:"timestamp"  validate:"required"`
}

type GpsPointResponse struct {
	ID        int       `json:"id" validate:"required"`
	GpsID     int       `json:"gps_id" validate:"required"`
	Bearing   float64   `json:"bearing" validate:"required"`
	Latitude  float64   `json:"latitude" validate:"required"`
	Longitude float64   `json:"longitude" validate:"required"`
	Timestamp int64     `json:"timestamp" validate:"required"`
	CreatedAt time.Time `json:"created_at" validate:"required"`
}
