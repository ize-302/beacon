// Package gpspoints
package gpspoints

import (
	"encoding/json"
	"net/http"

	"github.com/ize-302/beacon/backend/internal/database"

	_ "embed"
)

//go:embed queries/insert_gpspoint.sql
var insertGpsPoint string

//go:embed queries/select_gpspoints.sql
var selectGpsPoints string

type Handler struct {
	*database.Handler
	Hub interface {
		Broadcast(CreateGpsPoint)
	}
}

func (h *Handler) SaveGpsPoint(w http.ResponseWriter, r *http.Request) {
	var createGpsPoint CreateGpsPoint
	gpspoint := GpsPointResponse{}

	err := json.NewDecoder(r.Body).Decode(&createGpsPoint)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if createGpsPoint.GpsID == 0 {
		http.Error(w, "gps is required", http.StatusBadRequest)
		return
	}

	if createGpsPoint.Longitude == 0 {
		http.Error(w, "longitude is required", http.StatusBadRequest)
		return
	}

	if createGpsPoint.Latitude == 0 {
		http.Error(w, "latitude is required", http.StatusBadRequest)
		return
	}

	err = h.DB.QueryRow(insertGpsPoint, createGpsPoint.GpsID, createGpsPoint.Latitude, createGpsPoint.Longitude).Scan(&gpspoint.ID, &gpspoint.GpsID, &gpspoint.Latitude, &gpspoint.Longitude, &gpspoint.CreatedAt)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	// response := &GpsPointResponse{
	// 	ID:        gpspoint.ID,
	// 	GpsID:     gpspoint.GpsID,
	// 	Latitude:  gpspoint.Latitude,
	// 	Longitude: gpspoint.Longitude,
	// 	CreatedAt: gpspoint.CreatedAt,
	// }
	json.NewEncoder(w).Encode(gpspoint)

	// broadcast position
	if h.Hub != nil {
		h.Hub.Broadcast(createGpsPoint)
	}
}

func (h *Handler) FetchGpsPoints(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var gpspoints []GpsPointResponse

	rows, err := h.DB.Query(selectGpsPoints)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var gpspoint GpsPointResponse

		err = rows.Scan(&gpspoint.ID, &gpspoint.GpsID, &gpspoint.Latitude, &gpspoint.Longitude, &gpspoint.CreatedAt)
		if err != nil {
			panic(err)
		}

		gpspoints = append(gpspoints, gpspoint)
	}
	err = rows.Err()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(gpspoints)
}
