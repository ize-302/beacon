package main

import (
	"context"
	"fmt"
	"log"
	"os"

	movementengine "github.com/ize-302/beacon/backend/cmd/simulator/movement_engine"
	"github.com/ize-302/osmgraph/osmgraph"
	"github.com/joho/godotenv"
)

var baseURL string

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using environment variables")
	}
	baseURL = os.Getenv("API_BASE_URL")
	if baseURL == "" {
		baseURL = fmt.Sprintf("http://localhost:%s", os.Getenv("PORT"))
	}

	f, err := os.Open("cmd/simulator/map_data/lagos.osm.pbf")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// this now uses an osm library by yours truely: https://github.com/ize-302/osmgraph
	nodes, adj, err := osmgraph.GraphBuilder(f, osmgraph.DefaultRoadFilter, osmgraph.DefaultOneway)
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	movementengine.Run(baseURL, nodes, adj, ctx)
}
