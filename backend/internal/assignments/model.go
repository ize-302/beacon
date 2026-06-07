package assignments

import "time"

type Assignment struct {
	ID        int
	VehicleID int
	RiderID   int
	CreatedAt time.Time
}
