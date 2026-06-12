package main

import (
	"context"
	"fmt"
	"log"
	"os"

	graphbuilder "github.com/ize-302/beacon/backend/cmd/simulator/graph_builder"
	movementengine "github.com/ize-302/beacon/backend/cmd/simulator/movement_engine"
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

	nodes, adj := graphbuilder.BuildGraph()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	movementengine.StartSimulation(baseURL, nodes, adj, ctx)
}
