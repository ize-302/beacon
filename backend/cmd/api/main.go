package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/ize-302/beacon/backend/internal/common"
	"github.com/ize-302/beacon/backend/internal/database"
	"github.com/ize-302/beacon/backend/internal/health"
	"github.com/ize-302/beacon/backend/internal/vehicles"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using environment variables")
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
	_ = apiGroup

	// health routes
	health.NewHealthHander(apiGroup).RegisterRoutes()

	// vehicles
	vehicleRepo := vehicles.NewVehicleRepository(db)
	vehicleService := vehicles.NewVehicleService(vehicleRepo)
	vehicles.NewVehicleHandler(apiGroup, vehicleService).RegisterRoutes()

	// gps devices

	// gps-points

	// websocket

	appPort := os.Getenv("PORT")
	if appPort == "" {
		appPort = "8080"
	}
	fmt.Printf("Server listening on port %s...\n", appPort)
	err = http.ListenAndServe("127.0.0.1:"+appPort, router)
	if err != nil {
		fmt.Printf("Server failed to listen on port %s\n", appPort)
	}
}
