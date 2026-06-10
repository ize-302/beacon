package gps

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	_ "embed"

	"github.com/ize-302/beacon/backend/internal/database"
	"github.com/ize-302/beacon/backend/internal/vehicles"
)

//go:embed queries/insert_gps.sql
var insertGps string

//go:embed queries/select_gpss.sql
var selectGpss string

//go:embed queries/select_gps.sql
var selectGps string

//go:embed queries/delete_gps.sql
var deleteGps string

type Handler struct {
	*database.Handler
}

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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(gps)
}

func (h *Handler) FetchGpss(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var gpss []GpsResponse

	rows, err := h.DB.Query(selectGpss)
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
		gpss = append(gpss, gps)
	}
	err = rows.Err()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(gpss)
}

func (h *Handler) getGps(id int) *sql.Row {
	row := h.DB.QueryRow(selectGps, id)
	return row
}

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
