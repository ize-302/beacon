package vehicles

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/ize-302/beacon/backend/internal/database"

	_ "embed"
)

//go:embed queries/insert_vehicle.sql
var insertVehicle string

//go:embed queries/select_vehicles.sql
var selectVehicles string

//go:embed queries/select_vehicle.sql
var selectVehicle string

//go:embed queries/delete_vehicle.sql
var deleteVehicle string

type Handler struct {
	*database.Handler
}

// @Summary      Create a vehicle
// @Tags         vehicles
// @Accept       json
// @Produce      json
// @Param        body body CreateVehicleRequest true "Vehicle payload"
// @Success      201 {object} VehicleResponse
// @Failure      400 {string} string
// @Router       /vehicles [post]
func (h *Handler) CreateVehicle(w http.ResponseWriter, r *http.Request) {
	var createVehicleRequest CreateVehicleRequest
	vehicle := Vehicle{}
	err := json.NewDecoder(r.Body).Decode(&createVehicleRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if createVehicleRequest.PlateNumber == "" {
		http.Error(w, "plate_number is required", http.StatusBadRequest)
		return
	}

	err = h.DB.QueryRow(insertVehicle, createVehicleRequest.PlateNumber).Scan(&vehicle.ID, &vehicle.PlateNumber, &vehicle.CreatedAt)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	response := &VehicleResponse{
		ID:          vehicle.ID,
		PlateNumber: vehicle.PlateNumber,
		CreatedAt:   vehicle.CreatedAt,
	}
	json.NewEncoder(w).Encode(response)
}

// @Summary      List vehicles
// @Tags         vehicles
// @Produce      json
// @Success      200 {array} VehicleResponse
// @Router       /vehicles [get]
func (h *Handler) FetchVehicles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var vehicles []VehicleResponse

	rows, err := h.DB.Query(selectVehicles)
	if err != nil {
		panic(err)
	}

	defer rows.Close()
	for rows.Next() {
		var vehicle VehicleResponse
		err = rows.Scan(&vehicle.ID, &vehicle.PlateNumber, &vehicle.CreatedAt)
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
	row := h.DB.QueryRow(selectVehicle, id)
	return row
}

// @Summary      Get a vehicle
// @Tags         vehicles
// @Produce      json
// @Param        id path int true "Vehicle ID"
// @Success      200 {object} VehicleResponse
// @Failure      404 {string} string
// @Router       /vehicles/{id} [get]
func (h *Handler) FetchVehicle(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		panic(err)
	}

	var vehicle Vehicle
	row := h.getVehicleByID(id)
	switch err := row.Scan(&vehicle.ID, &vehicle.PlateNumber, &vehicle.CreatedAt); err {
	case sql.ErrNoRows:
		http.Error(w, "vehicle not found", http.StatusNotFound)
	case nil:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		response := &VehicleResponse{
			ID:          vehicle.ID,
			PlateNumber: vehicle.PlateNumber,
			CreatedAt:   vehicle.CreatedAt,
		}
		json.NewEncoder(w).Encode(response)
	default:
		panic(err)
	}
}

// @Summary      Delete a vehicle
// @Tags         vehicles
// @Param        id path int true "Vehicle ID"
// @Success      204
// @Failure      404 {string} string
// @Router       /vehicles/{id} [delete]
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
		_ = h.DB.QueryRow(deleteVehicle, id)
		w.WriteHeader(http.StatusNoContent)
	default:
		panic(err)
	}
}
