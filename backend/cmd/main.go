package main

import (
	"fmt"
	"net/http"

	"github.com/ize-302/beacon/backend/configs"
	_ "github.com/lib/pq"
)

func main() {
	var err error
	db, err = configs.DBConn()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	mux := http.NewServeMux()

	// riders
	mux.HandleFunc("POST /riders", CreateRider)

	mux.HandleFunc("GET /riders", FetchRiders)

	mux.HandleFunc("GET /riders/{id}", FetchRider)

	mux.HandleFunc("DELETE /riders/{id}", DeleteRider)

	// vehicles
	mux.HandleFunc("POST /vehicles", CreateVehicle)

	mux.HandleFunc("GET /vehicles", FetchVehicles)

	mux.HandleFunc("GET /vehicles/{id}", FetchVehicle)

	mux.HandleFunc("DELETE /vehicles/{id}", DeleteVehicle)

	// assignments
	mux.HandleFunc("POST /vehicle-assignments", AssignRiderToVehicle)

	mux.HandleFunc("GET /vehicle-assignments", FetchAssignments)

	mux.HandleFunc("GET /vehicle-assignments/{id}", FetchAssignment)

	mux.HandleFunc("DELETE /vehicle-assignments/{id}", UnassignRiderFromVehicle)

	// auditlogs
	// mux.HandleFunc("GET /logs", FetchLogs)

	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		fmt.Println("Server failed to listen on port 8080...")
		return
	}
	fmt.Println("Server listening on port 8080...")
}
