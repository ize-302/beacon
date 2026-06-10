package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/ize-302/beacon/backend/cmd/simulator/apis"
	graphbuilder "github.com/ize-302/beacon/backend/cmd/simulator/graph_builder"
	movementengine "github.com/ize-302/beacon/backend/cmd/simulator/movement_engine"
	internalgps "github.com/ize-302/beacon/backend/internal/gps"
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

	gpss, err := apis.APIFetchGpss(baseURL)
	if err != nil {
		log.Fatalf("Error parsing JSON: %v", err)
	}

	nodes, adj := graphbuilder.BuildGraph()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	for _, gps := range gpss {
		wg.Add(1)
		go func(gps internalgps.GpsResponse) {
			defer wg.Done()
			// fmt.Printf("Gps %d nodes: %d, adj nodes: %d\n", gps.ID, len(nodes), len(adj))
			movementengine.StartSimulation(baseURL, gps, nodes, adj, ctx)
		}(gps)
	}
	wg.Wait()
}
