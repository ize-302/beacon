// Package movementengine
package movementengine

import (
	"context"
	"math"
	"math/rand"
	"time"

	"github.com/ize-302/beacon/backend/cmd/simulator/apis"
	"github.com/ize-302/beacon/backend/cmd/simulator/models"
	gpspoints "github.com/ize-302/beacon/backend/internal/gps-points"

	internalgps "github.com/ize-302/beacon/backend/internal/gps"
)

func pickRandomNode(adj map[int64][]int64) int64 {
	keys := make([]int64, 0, len(adj))
	for k := range adj {
		keys = append(keys, k)
	}
	return keys[rand.Intn(len(keys))]
}

func closestNode(nodes map[int64]models.Node, lat, lng float64) int64 {
	var closest int64
	minDist := math.MaxFloat64
	for id, n := range nodes {
		latDiff := n.Lat - lat
		lngDiff := n.Lng - lng
		d := (latDiff * latDiff) + (lngDiff * lngDiff)
		if d < minDist {
			minDist = d
			closest = id
		}
	}
	return closest
}

func StartSimulation(baseURL string, gps internalgps.GpsResponse, nodes map[int64]models.Node, adj map[int64][]int64, ctx context.Context) {
	var current int64
	if gps.LastCoordinate != nil {
		current = closestNode(nodes, gps.LastCoordinate.Latitude, gps.LastCoordinate.Longitude)
	} else {
		current = pickRandomNode(adj)
	}
	randomSpeed := rand.Intn(9) + 1
	t := time.NewTicker(time.Duration(randomSpeed) * time.Second)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			neighbors := adj[current]
			if len(neighbors) == 0 {
				current = pickRandomNode(adj)
				continue
			}
			next := neighbors[rand.Intn(len(neighbors))]
			current = next

			node, ok := nodes[current]
			if !ok {
				continue
			}
			apis.APISendGpsPosition(gpspoints.CreateGpsPoint{
				GpsID:     gps.ID,
				Latitude:  node.Lat,
				Longitude: node.Lng,
			}, baseURL)
		}
	}
}
