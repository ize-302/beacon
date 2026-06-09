// Package gpspoints
package gpspoints

import "time"

type GpsPoint struct {
	ID        int
	GpsID     int
	Latitude  float64
	Longitude float64
	CreatedAt time.Time
}
