// Package internal
package internal

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/ize-302/beacon/backend/models"
)

type Handler struct {
	DB *sql.DB
}

func (h *Handler) CreateRider(w http.ResponseWriter, r *http.Request) {
	var rider models.Rider
	err := json.NewDecoder(r.Body).Decode(&rider)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if rider.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	sqlStatement := `INSERT INTO riders (name) VALUES ($1) RETURNING id, name;`
	err = h.DB.QueryRow(sqlStatement, rider.Name).Scan(&rider.ID, &rider.Name)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(rider)
}

func (h *Handler) FetchRiders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var riders []models.Rider

	sqlStatement := `SELECT r.id, r.name FROM riders r;`
	rows, err := h.DB.Query(sqlStatement)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var rider models.Rider

		err = rows.Scan(&rider.ID, &rider.Name)
		if err != nil {
			panic(err)
		}

		riders = append(riders, rider)
	}
	err = rows.Err()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(riders)
}

func (h *Handler) getRider(id int) *sql.Row {
	sqlStatement := `SELECT id, name FROM riders WHERE id = $1;`
	row := h.DB.QueryRow(sqlStatement, id)
	return row
}

func (h *Handler) FetchRider(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		panic(err)
	}

	var rider models.Rider
	row := h.getRider(id)
	switch err := row.Scan(&rider.ID, &rider.Name); err {
	case sql.ErrNoRows:
		http.Error(w, "rider not found", http.StatusNotFound)
	case nil:
		w.Header().Set("Content-Type", "application/json")

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(rider)
	default:
		panic(err)
	}
}

func (h *Handler) DeleteRider(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	row := h.getRider(id)
	switch err := row.Scan(&id); err {
	case sql.ErrNoRows:
		http.Error(w, "user not found", http.StatusNotFound)
	case nil:
		w.Header().Set("Content-Type", "application/json")
		sqlStatement := `DELETE FROM riders WHERE id = $1`
		_ = h.DB.QueryRow(sqlStatement, id)
		w.WriteHeader(http.StatusNoContent)
	default:
		panic(err)
	}
}

func (h *Handler) CreateVehicle(w http.ResponseWriter, r *http.Request) {
	var vehicle models.Vehicle
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
	var vehicles []models.Vehicle

	sqlStatement := `SELECT id, plate_number FROM vehicles;`
	rows, err := h.DB.Query(sqlStatement)
	if err != nil {
		panic(err)
	}

	defer rows.Close()
	for rows.Next() {
		var vehicle models.Vehicle
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

	var vehicle models.Vehicle
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

func (h *Handler) AssignRiderToVehicle(w http.ResponseWriter, r *http.Request) {
	var assignment models.AssignmentRequest
	var vehicle models.Vehicle
	var rider models.Rider
	err := json.NewDecoder(r.Body).Decode(&assignment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sqlStatement := `INSERT INTO assignments (vehicle_id, rider_id) VALUES ($1, $2) RETURNING id, vehicle_id, rider_id;`
	err = h.DB.QueryRow(sqlStatement, assignment.VehicleID, assignment.RiderID).Scan(&assignment.ID, &vehicle.ID, &rider.ID)
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(assignment)
}

func (h *Handler) FetchAssignments(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var assignments []models.AssignmentResponse

	query := `
	SELECT 
		a.id AS assignment_id,
		v.id AS vehicle_id, 
		v.plate_number, 
		r.id AS rider_id, 
		r.name AS rider_name
	FROM assignments a
	INNER JOIN vehicles v ON a.vehicle_id = v.id
	INNER JOIN riders r ON a.rider_id = r.id
	`
	rows, err := h.DB.Query(query)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var assignment models.AssignmentResponse
		assignment.Vehicle = &models.Vehicle{}
		assignment.Rider = &models.Rider{}

		err = rows.Scan(&assignment.ID, &assignment.Vehicle.ID, &assignment.Vehicle.PlateNumber, &assignment.Rider.ID, &assignment.Rider.Name)
		if err != nil {
			panic(err)
		}

		assignments = append(assignments, assignment)
	}
	err = rows.Err()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(assignments)
}

func (h *Handler) getAssignmentByID(id int) *sql.Row {
	query := `
	SELECT 
		a.id AS assignment_id,
		v.id AS vehicle_id,
		v.plate_number,
		r.id AS rider_id,
		r.name AS rider_name
	FROM assignments a
	INNER JOIN vehicles v ON a.vehicle_id = v.id
	INNER JOIN riders r ON a.rider_id = r.id
	WHERE a.id = $1;
	`
	row := h.DB.QueryRow(query, id)
	return row
}

func (h *Handler) FetchAssignment(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		panic(err)
	}

	var assignment models.AssignmentResponse
	assignment.Vehicle = &models.Vehicle{}
	assignment.Rider = &models.Rider{}

	row := h.getAssignmentByID(id)

	switch err = row.Scan(&assignment.ID, &assignment.Vehicle.ID, &assignment.Vehicle.PlateNumber, &assignment.Rider.ID, &assignment.Rider.Name); err {
	case sql.ErrNoRows:
		http.Error(w, "assignment not found", http.StatusNotFound)
	case nil:
		w.Header().Set("Content-Type", "application/json")

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(assignment)
	default:
		panic(err)
	}
}

func (h *Handler) UnassignRiderFromVehicle(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var assignment models.AssignmentResponse

	row := h.getAssignmentByID(id)
	assignment.Vehicle = &models.Vehicle{}
	assignment.Rider = &models.Rider{}

	switch err := row.Scan(&assignment.ID, &assignment.Vehicle.ID, &assignment.Vehicle.PlateNumber, &assignment.Rider.ID, &assignment.Rider.Name); err {
	case sql.ErrNoRows:
		http.Error(w, "assignment not found", http.StatusNotFound)
	case nil:
		w.Header().Set("Content-Type", "application/json")
		sqlStatement := `DELETE FROM assignments WHERE id = $1`
		_ = h.DB.QueryRow(sqlStatement, id)
		w.WriteHeader(http.StatusNoContent)
	default:
		panic(err)
	}
}
