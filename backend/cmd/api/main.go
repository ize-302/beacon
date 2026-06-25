package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/ize-302/beacon/backend/internal/common"
	"github.com/ize-302/beacon/backend/internal/database"
	"github.com/ize-302/beacon/backend/internal/gps"
	gpspoints "github.com/ize-302/beacon/backend/internal/gps-points"
	"github.com/ize-302/beacon/backend/internal/health"
	"github.com/ize-302/beacon/backend/internal/vehicles"
	"github.com/ize-302/beacon/backend/internal/ws"
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
	db, err := database.DBConn()
	if err != nil {
		panic(err)
	}
	fmt.Println("successfully connected!")
	defer db.Close()

	h := &database.Handler{DB: db}
	if err = h.SeedDB(); err != nil {
		fmt.Println("Error occured while seeding db", err)
	} else {
		fmt.Println("successfully seeded database")
	}

	router := chi.NewMux()

	// huma specific configs
	config := huma.DefaultConfig("Beacon API", "1.0.0")
	config.Info.Description = "Real-time vehicle tracking API"
	config.DocsRenderer = huma.DocsRendererSwaggerUI
	config.CreateHooks = nil // disabled $schema
	config.DocsPath = "/swagger"
	huma.NewError = func(status int, message string, errs ...error) huma.StatusError {
		return &common.MyError{
			Data:    nil,
			Status:  status,
			Message: message,
		}
	}

	api := humachi.New(router, config)
	apiGroup := huma.NewGroup(api, "/api/v1")

	// health routes
	health.NewHealthHander(apiGroup).RegisterRoutes()

	// vehicles
	vehicleRepo := vehicles.NewVehicleRepository(db)
	vehicleService := vehicles.NewVehicleService(vehicleRepo)
	vehicles.NewVehicleHandler(apiGroup, vehicleService).RegisterRoutes()

	// gps devices
	gpsRepo := gps.NewGpsRepository(db)
	eventHub := gps.NewEventHub()
	gpsService := gps.NewGpsService(gpsRepo, eventHub)
	gps.NewGpsHandler(apiGroup, gpsService, router).RegisterRoutes()

	// gps-points
	wsHub := ws.NewHub()
	gpsPointRepo := gpspoints.NewGpsPointRepository(db)
	gpsPointService := gpspoints.NewGpsPointService(gpsPointRepo, wsHub)
	gpspoints.NewGpsPointHandler(apiGroup, gpsPointService).RegisterRoutes()

	// websocket
	ws.NewWsHandler(wsHub).RegisterRoutes(router)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("Server listening on port %s...\n", port)
	err = http.ListenAndServe("127.0.0.1:"+port, corsMiddleware(router))
	if err != nil {
		fmt.Printf("Server failed to listen on port %s\n", port)
	}
}
