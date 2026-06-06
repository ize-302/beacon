package vehicles

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/ize-302/beacon/backend/internal/database"
)

type Handler struct {
	*database.Handler
}

func (h *Handler) CreateVehicle(w http.ResponseWriter, r *http.Request) {
	var vehicle Vehicle
	err := json.NewDecoder(r.Body).Decode(&vehicle)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if vehicle.PlateNumber == "" {
		http.Error(w, "plate_number is required", http.StatusBadRequest)
		return
	}

	sqlStatement := `INSERT INTO vehicles (plate_number) VALUES ($1) RETURNING id, plate_number;`
	err = h.DB.QueryRow(sqlStatement, vehicle.PlateNumber).Scan(&vehicle.ID, &vehicle.PlateNumber)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(vehicle)
}

func (h *Handler) FetchVehicles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var vehicles []Vehicle

	sqlStatement := `SELECT id, plate_number FROM vehicles;`
	rows, err := h.DB.Query(sqlStatement)
	if err != nil {
		panic(err)
	}

	defer rows.Close()
	for rows.Next() {
		var vehicle Vehicle
		err = rows.Scan(&vehicle.ID, &vehicle.PlateNumber)
		if err != nil {
			panic(err)
		}
		vehicles = append(vehicles, vehicle)
	}
	err = rows.Err()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(vehicles)
}

func (h *Handler) getVehicleByID(id int) *sql.Row {
	sqlStatement := `SELECT id, plate_number FROM vehicles WHERE id = $1;`
	row := h.DB.QueryRow(sqlStatement, id)
	return row
}

func (h *Handler) FetchVehicle(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		panic(err)
	}

	var vehicle Vehicle
	row := h.getVehicleByID(id)
	switch err := row.Scan(&vehicle.ID, &vehicle.PlateNumber); err {
	case sql.ErrNoRows:
		http.Error(w, "vehicle not found", http.StatusNotFound)
	case nil:
		w.Header().Set("Content-Type", "application/json")

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(vehicle)
	default:
		panic(err)
	}
}

func (h *Handler) DeleteVehicle(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	row := h.getVehicleByID(id)
	switch err := row.Scan(&id); err {
	case sql.ErrNoRows:
		http.Error(w, "vehicle not found", http.StatusNotFound)
	case nil:
		w.Header().Set("Content-Type", "application/json")
		sqlStatement := `DELETE FROM vehicles WHERE id = $1`
		_ = h.DB.QueryRow(sqlStatement, id)
		w.WriteHeader(http.StatusNoContent)
	default:
		panic(err)
	}
}
