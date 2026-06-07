// Package locations
package locations

import (
	"encoding/json"
	"net/http"

	"github.com/ize-302/beacon/backend/internal/database"
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
	location.Vehicle = &vehicles.VehicleResponse{}
	err := json.NewDecoder(r.Body).Decode(&createLocation)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if createLocation.VehicleID == 0 {
		http.Error(w, "vehicle_id is required", http.StatusBadRequest)
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

	err = h.DB.QueryRow(insertLocation, createLocation.VehicleID, createLocation.Latitude, createLocation.Longitude).Scan(&location.ID, &location.Latitude, &location.Longitude, &location.CreatedAt, &location.Vehicle.ID, &location.Vehicle.PlateNumber, &location.Vehicle.CreatedAt)
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
		Vehicle: &vehicles.VehicleResponse{
			ID:          location.Vehicle.ID,
			PlateNumber: location.Vehicle.PlateNumber,
			CreatedAt:   location.Vehicle.CreatedAt,
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
		location.Vehicle = &vehicles.VehicleResponse{}

		err = rows.Scan(&location.ID, &location.Latitude, &location.Longitude, &location.CreatedAt, &location.Vehicle.ID, &location.Vehicle.PlateNumber, &location.Vehicle.CreatedAt)
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
