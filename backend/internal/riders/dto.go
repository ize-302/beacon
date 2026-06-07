// Package riders
package riders

import "time"

type CreateRiderRequest struct {
	Name string `json:"name"`
}

type RiderResponse struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}
