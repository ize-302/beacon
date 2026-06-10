// Package gpspoints
package gpspoints

import (
	"time"
)

type CreateGpsPoint struct {
	GpsID     int     `json:"gps_id"`
	Bearing   float64 `json:"bearing"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timestamp int64   `json:"timestamp"`
}

type GpsPointResponse struct {
	ID        int       `json:"id"`
	GpsID     int       `json:"gps_id"`
	Bearing   float64   `json:"bearing"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Timestamp int64     `json:"timestamp"`
	CreatedAt time.Time `json:"created_at"`
}
