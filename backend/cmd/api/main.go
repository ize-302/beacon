package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/ize-302/beacon/backend/internal"
	"github.com/ize-302/beacon/backend/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	var err error

	err = godotenv.Load()
	if err != nil {
		log.Fatal("")
	}

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	conn := fmt.Sprintf("host=%s port=%s user=%s  password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	db, err := database.DBConn(conn)
	if err != nil {
		panic(err)
	}
	fmt.Println("successfully connected!")
	defer db.Close()

	mux := http.NewServeMux()

	h := &internal.Handler{
		DB: db,
	}

	// riders
	mux.HandleFunc("POST /riders", h.CreateRider)

	mux.HandleFunc("GET /riders", h.FetchRiders)

	mux.HandleFunc("GET /riders/{id}", h.FetchRider)

	mux.HandleFunc("DELETE /riders/{id}", h.DeleteRider)

	// vehicles
	mux.HandleFunc("POST /vehicles", h.CreateVehicle)

	mux.HandleFunc("GET /vehicles", h.FetchVehicles)

	mux.HandleFunc("GET /vehicles/{id}", h.FetchVehicle)

	mux.HandleFunc("DELETE /vehicles/{id}", h.DeleteVehicle)

	// assignments
	mux.HandleFunc("POST /vehicle-assignments", h.AssignRiderToVehicle)

	mux.HandleFunc("GET /vehicle-assignments", h.FetchAssignments)

	mux.HandleFunc("GET /vehicle-assignments/{id}", h.FetchAssignment)

	mux.HandleFunc("DELETE /vehicle-assignments/{id}", h.UnassignRiderFromVehicle)

	// auditlogs
	// mux.HandleFunc("GET /logs", FetchLogs)

	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		fmt.Println("Server failed to listen on port 8080...")
		return
	}
	fmt.Println("Server listening on port 8080...")
}
