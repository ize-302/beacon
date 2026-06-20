// Package gpspoints
package gpspoints

import "time"

type CreateGpsPoint struct {
	GpsID     int     `json:"gps_id" validate:"required"`
	Bearing   float64 `json:"bearing"`
	Latitude  float64 `json:"latitude" validate:"required"`
	Longitude float64 `json:"longitude" validate:"required"`
	Timestamp int64   `json:"timestamp" validate:"required"`
}

type CreateGpsPointRequest struct {
	Body *CreateGpsPoint
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
