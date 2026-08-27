package main

import (
	"context"
	"log"
	"os"

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

	baseURL := os.Getenv("API_BASE_URL")

	nodes, adj, err := sim.LoadGraph(mapDataPath)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sim.Run(baseURL, nodes, adj, ctx)
}
