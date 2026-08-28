// Package gpspoints
package gpspoints

import "time"

type GpsPoint struct {
	ID        int
	VehicleID int
	Bearing   float64
	Latitude  float64
	Longitude float64
	Timestamp int64
	CreatedAt time.Time
}
