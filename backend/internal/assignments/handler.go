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

	_ "embed"
)

//go:embed queries/insert_assignment.sql
var insertAssignment string

//go:embed queries/select_assignments.sql
var selectAssignments string

type Handler struct {
	*database.Handler
}

func (h *Handler) AssignRiderToVehicle(w http.ResponseWriter, r *http.Request) {
	var createAssignmentRequest CreateAssignmentRequest
	assignment := AssignmentResponse{}
	assignment.Vehicle = &vehicles.VehicleResponse{}
	assignment.Rider = &riders.RiderResponse{}

	err := json.NewDecoder(r.Body).Decode(&createAssignmentRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.DB.QueryRow(insertAssignment, createAssignmentRequest.VehicleID, createAssignmentRequest.RiderID).Scan(&assignment.ID, &assignment.Vehicle.ID, &assignment.Vehicle.PlateNumber, &assignment.Rider.ID, &assignment.Rider.Name)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	response := &AssignmentResponse{
		ID: assignment.ID,
		Vehicle: &vehicles.VehicleResponse{
			ID:          assignment.Vehicle.ID,
			PlateNumber: assignment.Vehicle.PlateNumber,
		},
		Rider: &riders.RiderResponse{
			ID:   assignment.Rider.ID,
			Name: assignment.Rider.Name,
		},
	}
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) FetchAssignments(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var assignments []AssignmentResponse

	rows, err := h.DB.Query(selectAssignments)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var assignment AssignmentResponse
		assignment.Vehicle = &vehicles.VehicleResponse{}
		assignment.Rider = &riders.RiderResponse{}

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
	assignment.Vehicle = &vehicles.VehicleResponse{}
	assignment.Rider = &riders.RiderResponse{}

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
	assignment.Vehicle = &vehicles.VehicleResponse{}
	assignment.Rider = &riders.RiderResponse{}

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
