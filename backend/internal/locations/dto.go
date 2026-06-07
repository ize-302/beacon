// Package locations
package locations

import (
	"time"

	"github.com/ize-302/beacon/backend/internal/assignments"
)

type CreateLocation struct {
	AssignmentID int     `json:"assignment_id"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
}

type LocationResponse struct {
	ID         int                             `json:"id"`
	Assignment *assignments.AssignmentResponse `json:"assignment"`
	Latitude   float64                         `json:"latitude"`
	Longitude  float64                         `json:"longitude"`
	CreatedAt  time.Time                       `json:"created_at"`
}
