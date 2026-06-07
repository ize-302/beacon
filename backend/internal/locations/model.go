// Package locations
package locations

import "time"

type Location struct {
	ID        int
	VehicleID int
	Latitude  float64
	Longitude float64
	CreatedAt time.Time
}
