// Package assignments
package assignments

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/ize-302/beacon/backend/internal/database"
	"github.com/ize-302/beacon/backend/internal/riders"
	"github.com/ize-302/beacon/backend/internal/vehicles"
)

type Handler struct {
	*database.Handler
}

func (h *Handler) AssignRiderToVehicle(w http.ResponseWriter, r *http.Request) {
	var assignment AssignmentRequest
	var vehicle vehicles.Vehicle
	var rider riders.Rider
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
	var assignments []AssignmentResponse

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
		var assignment AssignmentResponse
		assignment.Vehicle = &vehicles.Vehicle{}
		assignment.Rider = &riders.Rider{}

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

	var assignment AssignmentResponse
	assignment.Vehicle = &vehicles.Vehicle{}
	assignment.Rider = &riders.Rider{}

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

	var assignment AssignmentResponse

	row := h.getAssignmentByID(id)
	assignment.Vehicle = &vehicles.Vehicle{}
	assignment.Rider = &riders.Rider{}

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
