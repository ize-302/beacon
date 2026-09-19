// Package vehicles
package vehicles

import "time"

type Vehicle struct {
	ID          int
	PlateNumber string
	CreatedAt   time.Time
}
