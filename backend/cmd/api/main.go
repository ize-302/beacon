package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/ize-302/beacon/backend/internal/assignments"
	"github.com/ize-302/beacon/backend/internal/database"
	"github.com/ize-302/beacon/backend/internal/locations"
	"github.com/ize-302/beacon/backend/internal/riders"
	"github.com/ize-302/beacon/backend/internal/vehicles"
	"github.com/ize-302/beacon/backend/internal/ws"
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

	h := &database.Handler{
		DB: db,
	}

	// SEED DATABASE
	err = h.SeedDB()
	if err != nil {
		fmt.Println("Error occured while seeding db", err)
		return
	} else {
		fmt.Println("successfully seeded database")
	}

	ridersHandler := &riders.Handler{Handler: h}

	vehiclesHandler := &vehicles.Handler{Handler: h}

	assignmentsHandler := &assignments.Handler{Handler: h}

	hub := ws.NewHub()
	locationsHandler := &locations.Handler{Handler: h, Hub: hub}

	socketHandler := &ws.Handler{Handler: h}

	// websocket
	mux.HandleFunc("/ws", socketHandler.WsHandler(hub))

	// riders
	mux.HandleFunc("POST /riders", ridersHandler.CreateRider)

	mux.HandleFunc("GET /riders", ridersHandler.FetchRiders)

	mux.HandleFunc("GET /riders/{id}", ridersHandler.FetchRider)

	mux.HandleFunc("DELETE /riders/{id}", ridersHandler.DeleteRider)

	// vehicles
	mux.HandleFunc("POST /vehicles", vehiclesHandler.CreateVehicle)

	mux.HandleFunc("GET /vehicles", vehiclesHandler.FetchVehicles)

	mux.HandleFunc("GET /vehicles/{id}", vehiclesHandler.FetchVehicle)

	mux.HandleFunc("DELETE /vehicles/{id}", vehiclesHandler.DeleteVehicle)

	// assignments
	mux.HandleFunc("POST /vehicle-assignments", assignmentsHandler.AssignRiderToVehicle)

	mux.HandleFunc("GET /vehicle-assignments", assignmentsHandler.FetchAssignments)

	mux.HandleFunc("GET /vehicle-assignments/{id}", assignmentsHandler.FetchAssignment)

	mux.HandleFunc("DELETE /vehicle-assignments/{id}", assignmentsHandler.UnassignRiderFromVehicle)

	// locations
	mux.HandleFunc("POST /locations", locationsHandler.SaveLocation)

	mux.HandleFunc("GET /locations", locationsHandler.FetchLocations)

	// auditlogs
	// mux.HandleFunc("GET /logs", FetchLogs)

	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		fmt.Println("Server failed to listen on port 8080...")
		return
	}
	fmt.Println("Server listening on port 8080...")
}
