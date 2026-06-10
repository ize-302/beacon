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

// Breadth-First Search: explores the graph layer by layer by finding the
// shortedt possible path to a destination from a given current position. It also makes it
// impossible to revisit a node because BFS marks nodes visited and never
// includes duplicates in the path
// Summary on how it works:
// 1. Finds the shortest path from current position to that destination
// 2. Walks that path one node per tick
// 3. When it arrives, picks a new random destination and repeats
// Learn more about BFS algorithm here: https://www.youtube.com/watch?v=HZ5YTanv5QE
func bfsPath(adj map[int64][]int64, start, goal int64) []int64 {
	if start == goal {
		return []int64{start}
	}
	prev := map[int64]int64{start: -1}
	queue := []int64{start}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, nb := range adj[cur] {
			if _, visited := prev[nb]; visited {
				continue
			}
			prev[nb] = cur
			if nb == goal {
				path := []int64{}
				for n := goal; n != -1; n = prev[n] {
					path = append([]int64{n}, path...)
				}
				return path
			}
			queue = append(queue, nb)
		}
	}
	return nil // no path was found
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

	var path []int64

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			// determine new path when current one is exhausted
			for len(path) == 0 {
				dest := pickRandomNode(adj)
				if dest == current {
					continue
				}
				path = bfsPath(adj, current, dest)
				if len(path) > 1 {
					path = path[1:] // drop current node
				} else {
					path = nil
				}
			}

			current = path[0]
			path = path[1:]

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
