// Package riders
package riders

type CreateRiderRequest struct {
	Name string `json:"name"`
}

type RiderResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
