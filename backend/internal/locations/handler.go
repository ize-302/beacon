// Package locations
package locations

import (
	"encoding/json"
	"net/http"

	"github.com/ize-302/beacon/backend/internal/assignments"
	"github.com/ize-302/beacon/backend/internal/database"
	"github.com/ize-302/beacon/backend/internal/riders"
	"github.com/ize-302/beacon/backend/internal/vehicles"

	_ "embed"
)

//go:embed queries/insert_location.sql
var insertLocation string

//go:embed queries/select_locations.sql
var selectLocations string

type Handler struct {
	*database.Handler
}

func (h *Handler) SaveLocation(w http.ResponseWriter, r *http.Request) {
	var createLocation CreateLocation
	location := LocationResponse{}
	location.Assignment = &assignments.AssignmentResponse{}
	location.Assignment.Vehicle = &vehicles.VehicleResponse{}
	location.Assignment.Rider = &riders.RiderResponse{}

	err := json.NewDecoder(r.Body).Decode(&createLocation)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if createLocation.AssignmentID == 0 {
		http.Error(w, "assignment_id is required", http.StatusBadRequest)
		return
	}

	if createLocation.Longitude == 0 {
		http.Error(w, "longitude is required", http.StatusBadRequest)
		return
	}

	if createLocation.Latitude == 0 {
		http.Error(w, "latitude is required", http.StatusBadRequest)
		return
	}

	err = h.DB.QueryRow(insertLocation, createLocation.AssignmentID, createLocation.Latitude, createLocation.Longitude).Scan(&location.ID, &location.Latitude, &location.Longitude, &location.CreatedAt, &location.Assignment.ID, &location.Assignment.CreatedAt, &location.Assignment.Vehicle.ID, &location.Assignment.Vehicle.PlateNumber, &location.Assignment.Vehicle.CreatedAt, &location.Assignment.Rider.ID, &location.Assignment.Rider.Name, &location.Assignment.Rider.CreatedAt)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	response := &LocationResponse{
		ID:        location.ID,
		Latitude:  location.Latitude,
		Longitude: location.Longitude,
		CreatedAt: location.CreatedAt,
		Assignment: &assignments.AssignmentResponse{
			Vehicle: &vehicles.VehicleResponse{
				ID:          location.Assignment.Vehicle.ID,
				PlateNumber: location.Assignment.Vehicle.PlateNumber,
				CreatedAt:   location.Assignment.Vehicle.CreatedAt,
			},
			Rider: &riders.RiderResponse{
				ID:        location.Assignment.Rider.ID,
				Name:      location.Assignment.Rider.Name,
				CreatedAt: location.Assignment.Rider.CreatedAt,
			},
		},
	}
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) FetchLocations(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var locations []LocationResponse

	rows, err := h.DB.Query(selectLocations)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var location LocationResponse
		location.Assignment = &assignments.AssignmentResponse{}
		location.Assignment.Vehicle = &vehicles.VehicleResponse{}
		location.Assignment.Rider = &riders.RiderResponse{}

		err = rows.Scan(&location.ID, &location.Latitude, &location.Longitude, &location.CreatedAt, &location.Assignment.ID, &location.Assignment.CreatedAt, &location.Assignment.Vehicle.ID, &location.Assignment.Vehicle.PlateNumber, &location.Assignment.Vehicle.CreatedAt, &location.Assignment.Rider.ID, &location.Assignment.Rider.Name, &location.Assignment.Rider.CreatedAt)
		if err != nil {
			panic(err)
		}

		locations = append(locations, location)
	}
	err = rows.Err()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(locations)
}
