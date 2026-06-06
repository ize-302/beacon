package assignments

import (
	"github.com/ize-302/beacon/backend/internal/riders"
	"github.com/ize-302/beacon/backend/internal/vehicles"
)

type CreateAssignmentRequest struct {
	ID        int `json:"id"`
	VehicleID int `json:"vehicle_id"`
	RiderID   int `json:"rider_id"`
}

type AssignmentResponse struct {
	ID      int                       `json:"id"`
	Vehicle *vehicles.VehicleResponse `json:"vehicle"`
	Rider   *riders.RiderResponse     `json:"rider"`
}
