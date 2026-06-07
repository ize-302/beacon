// Package locations
package locations

import (
	"encoding/json"
	"net/http"

	"github.com/ize-302/beacon/backend/internal/database"
	"github.com/ize-302/beacon/backend/internal/vehicles"
)

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

	sqlStatement := `
		WITH inserted AS (
			INSERT INTO locations (vehicle_id, latitude, longitude)
			VALUES ($1, $2, $3)
			RETURNING id, vehicle_id, latitude, longitude, created_at 
		)
		SELECT
			i.id,
			i.latitude,
			i.longitude,
			i.created_at,
			v.id,
			v.plate_number,
			v.created_at
		FROM inserted i
		JOIN vehicles v ON v.id = i.vehicle_id
	`

	err = h.DB.QueryRow(sqlStatement, createLocation.VehicleID, createLocation.Latitude, createLocation.Longitude).Scan(&location.ID, &location.Latitude, &location.Longitude, &location.CreatedAt, &location.Vehicle.ID, &location.Vehicle.PlateNumber, &location.Vehicle.CreatedAt)
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

	query := `
		SELECT 
			l.id AS location_id,
			l.latitude AS location_latitude,
			l.longitude AS location_longitude,
			l.created_at AS location_created_at,
			v.id AS vehicle_id, 
			v.plate_number, 
			v.created_at
		FROM locations l
		INNER JOIN vehicles v ON l.vehicle_id = v.id
	`
	rows, err := h.DB.Query(query)
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
