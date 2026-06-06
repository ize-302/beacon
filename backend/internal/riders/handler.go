package riders

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

	sqlStatement := `INSERT INTO riders (name) VALUES ($1) RETURNING id, name;`
	err = h.DB.QueryRow(sqlStatement, createRiderRequest.Name).Scan(&rider.ID, &rider.Name)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := &RiderResponse{
		ID:   rider.ID,
		Name: rider.Name,
	}
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) FetchRiders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var riders []RiderResponse

	sqlStatement := `SELECT r.id, r.name FROM riders r;`
	rows, err := h.DB.Query(sqlStatement)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var rider RiderResponse

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

	var rider Rider
	row := h.getRider(id)
	switch err := row.Scan(&rider.ID, &rider.Name); err {
	case sql.ErrNoRows:
		http.Error(w, "rider not found", http.StatusNotFound)
	case nil:
		w.Header().Set("Content-Type", "application/json")

		w.WriteHeader(http.StatusOK)
		response := &RiderResponse{
			ID:   rider.ID,
			Name: rider.Name,
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
		sqlStatement := `DELETE FROM riders WHERE id = $1`
		_ = h.DB.QueryRow(sqlStatement, id)
		w.WriteHeader(http.StatusNoContent)
	default:
		panic(err)
	}
}
