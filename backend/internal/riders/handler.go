package riders

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	_ "embed"

	"github.com/ize-302/beacon/backend/internal/database"
)

//go:embed queries/insert_rider.sql
var insertRider string

//go:embed queries/select_riders.sql
var selectRiders string

//go:embed queries/select_rider.sql
var selectRider string

//go:embed queries/delete_rider.sql
var deleteRider string

type Handler struct {
	*database.Handler
}

func (h *Handler) CreateRider(w http.ResponseWriter, r *http.Request) {
	var createRiderRequest CreateRiderRequest
	rider := Rider{}
	err := json.NewDecoder(r.Body).Decode(&createRiderRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if createRiderRequest.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	err = h.DB.QueryRow(insertRider, createRiderRequest.Name).Scan(&rider.ID, &rider.Name, &rider.CreatedAt)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := &RiderResponse{
		ID:        rider.ID,
		Name:      rider.Name,
		CreatedAt: rider.CreatedAt,
	}
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) FetchRiders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var riders []RiderResponse

	rows, err := h.DB.Query(selectRiders)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var rider RiderResponse

		err = rows.Scan(&rider.ID, &rider.Name, &rider.CreatedAt)
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
	row := h.DB.QueryRow(selectRider, id)
	return row
}

func (h *Handler) FetchRider(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		panic(err)
	}

	var rider Rider
	row := h.getRider(id)
	switch err := row.Scan(&rider.ID, &rider.Name, &rider.CreatedAt); err {
	case sql.ErrNoRows:
		http.Error(w, "rider not found", http.StatusNotFound)
	case nil:
		w.Header().Set("Content-Type", "application/json")

		w.WriteHeader(http.StatusOK)
		response := &RiderResponse{
			ID:        rider.ID,
			Name:      rider.Name,
			CreatedAt: rider.CreatedAt,
		}
		json.NewEncoder(w).Encode(response)
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
		_ = h.DB.QueryRow(deleteRider, id)
		w.WriteHeader(http.StatusNoContent)
	default:
		panic(err)
	}
}
