package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/ize-302/beacon/backend/internal/sim"
	"github.com/joho/godotenv"
)

// mapDataPath is resolved relative to the working directory, so the simulator
// must be run from backend/.
const mapDataPath = "cmd/simulator/map_data/lagos.osm.pbf"

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using environment variables!")
	}

	graph, err := sim.LoadGraph(mapDataPath)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("simulator: graph loaded (%d routable nodes)", graph.Size())

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg := sim.Config{
		BaseURL:   os.Getenv("API_BASE_URL"),
		Graph:     graph,
		Planners:  envInt("SIM_PLANNERS"),
		AdmitRate: envInt("SIM_ADMIT_RATE"),
		Seed:      int64(envInt("SIM_SEED")),
	}

	if err := sim.Run(ctx, cfg); err != nil {
		log.Fatal(err)
	}
}

// envInt returns 0 when the variable is unset, so the config default applies.
// A value that is set but not a number is a mistake, not a request for the default.
func envInt(key string) int {
	v := os.Getenv(key)
	if v == "" {
		return 0
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		log.Fatalf("%s must be an integer, got %q", key, v)
	}
	return n
}
