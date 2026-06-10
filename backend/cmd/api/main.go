package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/ize-302/beacon/backend/internal/database"
	"github.com/ize-302/beacon/backend/internal/gps"
	gpspoints "github.com/ize-302/beacon/backend/internal/gps-points"
	"github.com/ize-302/beacon/backend/internal/vehicles"
	"github.com/ize-302/beacon/backend/internal/ws"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

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

	gpsHandler := &gps.Handler{Handler: h}

	vehiclesHandler := &vehicles.Handler{Handler: h}

	// assignmentsHandler := &assignments.Handler{Handler: h}

	hub := ws.NewHub()
	gpspointsHandler := &gpspoints.Handler{Handler: h, Hub: hub}

	socketHandler := &ws.Handler{Handler: h}

	// websocket
	mux.HandleFunc("/ws", socketHandler.WsHandler(hub))

	// gpss
	mux.HandleFunc("POST /gps", gpsHandler.CreateGps)

	mux.HandleFunc("GET /gps", gpsHandler.FetchGpss)

	mux.HandleFunc("GET /gps/{id}", gpsHandler.FetchGps)

	mux.HandleFunc("DELETE /gps/{id}", gpsHandler.DeleteGps)

	// vehicles
	mux.HandleFunc("POST /vehicles", vehiclesHandler.CreateVehicle)

	mux.HandleFunc("GET /vehicles", vehiclesHandler.FetchVehicles)

	mux.HandleFunc("GET /vehicles/{id}", vehiclesHandler.FetchVehicle)

	mux.HandleFunc("DELETE /vehicles/{id}", vehiclesHandler.DeleteVehicle)

	// gps-points
	mux.HandleFunc("POST /gps-points", gpspointsHandler.SaveGpsPoint)

	mux.HandleFunc("GET /gps-points", gpspointsHandler.FetchGpsPoints)

	// auditlogs
	// mux.HandleFunc("GET /logs", FetchLogs)

	err = http.ListenAndServe(":8080", corsMiddleware(mux))
	if err != nil {
		fmt.Println("Server failed to listen on port 8080...")
		return
	}
	fmt.Println("Server listening on port 8080...")
}
