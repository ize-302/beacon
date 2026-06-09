// Package gps
package gps

import "time"

type Gps struct {
	ID        int
	SN        string
	VehicleID int
	CreatedAt time.Time
}
