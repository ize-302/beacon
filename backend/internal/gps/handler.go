package gps

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	_ "embed"

	"github.com/ize-302/beacon/backend/internal/database"
	"github.com/ize-302/beacon/backend/internal/vehicles"
)

//go:embed queries/insert_gps.sql
var insertGps string

//go:embed queries/select_gps_devices.sql
var selectGpsDevices string

//go:embed queries/select_gps.sql
var selectGps string

//go:embed queries/delete_gps.sql
var deleteGps string

//go:embed queries/select_gps_history.sql
var getGpsHistory string

type Handler struct {
	*database.Handler
	EventHub *EventHub
}

// @Summary      Register a GPS device
// @Tags         gps
// @Accept       json
// @Produce      json
// @Param        body body CreateGpsRequest true "GPS payload"
// @Success      201 {object} GpsResponse
// @Failure      400 {string} string
// @Router       /gps-devices [post]
func (h *Handler) CreateGps(w http.ResponseWriter, r *http.Request) {
	var createGpsRequest CreateGpsRequest
	gps := GpsResponse{}
	gps.Vehicle = &vehicles.VehicleResponse{}

	err := json.NewDecoder(r.Body).Decode(&createGpsRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if createGpsRequest.SN == "" {
		http.Error(w, "Serial number is required", http.StatusBadRequest)
		return
	}

	if createGpsRequest.VehicleID == 0 {
		http.Error(w, "Vehicle ID is required", http.StatusBadRequest)
		return
	}

	err = h.DB.QueryRow(insertGps, createGpsRequest.SN, createGpsRequest.VehicleID).Scan(&gps.ID, &gps.SN, &gps.CreatedAt, &gps.Vehicle.ID, &gps.Vehicle.PlateNumber, &gps.Vehicle.CreatedAt)
	if err != nil {
		panic(err)
	}

	h.EventHub.Publish(gps)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(gps)
}

// @Summary      Stream new GPS devices via SSE
// @Tags         gps
// @Produce      text/event-stream
// @Success      200
// @Router       /gps-devices/events [get]
func (h *Handler) StreamNewDevices(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	ch := h.EventHub.Subscribe()
	defer h.EventHub.Unsubscribe(ch)

	for {
		select {
		case <-r.Context().Done():
			return
		case gps := <-ch:
			data, err := json.Marshal(gps)
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}
	}
}

// @Summary      List GPS devices
// @Tags         gps
// @Produce      json
// @Success      200 {array} GpsResponse
// @Router       /gps-devices [get]
func (h *Handler) FetchGpsDevices(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var gpsDevices []GpsResponse

	rows, err := h.DB.Query(selectGpsDevices)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var gps GpsResponse
		gps.Vehicle = &vehicles.VehicleResponse{}

		var lat, lng sql.NullFloat64
		var lastAt sql.NullTime

		err = rows.Scan(&gps.ID, &gps.SN, &gps.CreatedAt, &gps.Vehicle.ID, &gps.Vehicle.PlateNumber, &gps.Vehicle.CreatedAt, &lat, &lng, &lastAt)
		if err != nil {
			panic(err)
		}
		if lat.Valid && lng.Valid {
			updatedAt := time.Time{}
			if lastAt.Valid {
				updatedAt = lastAt.Time
			}
			gps.LastCoordinate = &Coordinate{
				Latitude:  lat.Float64,
				Longitude: lng.Float64,
				UpdatedAt: updatedAt,
			}
		}
		gpsDevices = append(gpsDevices, gps)
	}
	err = rows.Err()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(gpsDevices)
}

func (h *Handler) getGps(id int) *sql.Row {
	row := h.DB.QueryRow(selectGps, id)
	return row
}

// @Summary      Get a GPS device
// @Tags         gps
// @Produce      json
// @Param        id path int true "GPS ID"
// @Success      200 {object} GpsResponse
// @Failure      404 {string} string
// @Router       /gps-devices/{id} [get]
func (h *Handler) FetchGps(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		panic(err)
	}

	var gps GpsResponse
	gps.Vehicle = &vehicles.VehicleResponse{}

	var lat, lng sql.NullFloat64
	var lastAt sql.NullTime

	row := h.getGps(id)
	switch err := row.Scan(&gps.ID, &gps.SN, &gps.CreatedAt, &gps.Vehicle.ID, &gps.Vehicle.PlateNumber, &gps.Vehicle.CreatedAt, &lat, &lng, &lastAt); err {
	case sql.ErrNoRows:
		http.Error(w, "gps not found", http.StatusNotFound)
	case nil:
		if lat.Valid && lng.Valid {
			updatedAt := time.Time{}
			if lastAt.Valid {
				updatedAt = lastAt.Time
			}
			gps.LastCoordinate = &Coordinate{
				Latitude:  lat.Float64,
				Longitude: lng.Float64,
				UpdatedAt: updatedAt,
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(gps)
	default:
		panic(err)
	}
}

// @Summary      Delete a GPS device
// @Tags         gps
// @Param        id path int true "GPS ID"
// @Success      204
// @Failure      404 {string} string
// @Router       /gps-devices/{id} [delete]
func (h *Handler) DeleteGps(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	row := h.getGps(id)
	switch err := row.Scan(&id); err {
	case sql.ErrNoRows:
		http.Error(w, "gps not found", http.StatusNotFound)
	case nil:
		w.Header().Set("Content-Type", "application/json")
		_ = h.DB.QueryRow(deleteGps, id)
		w.WriteHeader(http.StatusNoContent)
	default:
		panic(err)
	}
}

// @Summary      Get a GpsHistory ...
// @Tags         gps
// @Produce      json
// @Param        id path int true "GPS ID"
// @Success      200 {object} GpsHistoryResponse
// @Failure      404 {string} string
// @Router       /gps-devices/{id}/history [get]
func (h *Handler) GpsHistory(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	rows, err := h.DB.Query(getGpsHistory, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var gpsHistory GpsHistoryResponse
	coordinates := []Coordinate{}
	found := false

	for rows.Next() {
		var lat, lng sql.NullFloat64
		if err := rows.Scan(&gpsHistory.GpsID, &gpsHistory.GpsSN, &lat, &lng); err != nil {
			panic(err)
		}
		found = true
		if lat.Valid && lng.Valid {
			coordinates = append(coordinates, Coordinate{
				Latitude:  lat.Float64,
				Longitude: lng.Float64,
			})
		}
	}
	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !found {
		http.Error(w, "gps not found", http.StatusNotFound)
		return
	}

	gpsHistory.Coordinates = &coordinates
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(gpsHistory)
}
