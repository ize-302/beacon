// Package movementengine
package movementengine

import (
	"context"
	"log"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/ize-302/beacon/backend/cmd/simulator/apis"
	"github.com/ize-302/beacon/backend/cmd/simulator/models"
	gpspoints "github.com/ize-302/beacon/backend/internal/gps-points"

	internalgps "github.com/ize-302/beacon/backend/internal/gps"
)

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

// Using haversine bearing formula to compute gps bearing from one gps coordinate to another
// Formula:
// Δlng = lng2 − lng1
// x = sin(Δlng)·cos(lat2)
// y = cos(lat1)·sin(lat2) − sin(lat1)·cos(lat2)·cos(Δlng)
// bearing = atan2(x, y)           // radians, −π to +π
// degrees = (bearing·180/π + 360) % 360   // normalize to 0–360
// result: 0 = North, 90 = East, 180 = South, 270 = West.
func computeBearing(from, to models.Node) float64 {
	lat1 := from.Lat * math.Pi / 180
	lng1 := from.Lng * math.Pi / 180
	lat2 := to.Lat * math.Pi / 180
	lng2 := to.Lng * math.Pi / 180

	dLng := lng2 - lng1 // diff in longitue

	x := math.Sin(dLng) * math.Cos(lat2) // compute east-west
	y := math.Cos(lat1)*math.Sin(lat2) -
		math.Sin(lat1)*math.Cos(lat2)*math.Cos(dLng) // compute north-south

	bearing := math.Atan2(x, y)

	// Convert radians to degrees
	degrees := bearing * 180 / math.Pi
	// Normalize to 0-360
	degrees = math.Mod(degrees+360, 360)
	return degrees
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

func pickRandomNode(adj map[int64][]int64) int64 {
	keys := make([]int64, 0, len(adj))
	for k := range adj {
		keys = append(keys, k)
	}
	return keys[rand.Intn(len(keys))]
}

func DriveVehicle(baseURL string, gps internalgps.GpsResponse, nodes map[int64]models.Node, adj map[int64][]int64, ctx context.Context) {
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

			prevNode := nodes[current]
			current = path[0]
			path = path[1:]

			node, ok := nodes[current]
			if !ok {
				continue
			}
			apis.SendGpsPosition(gpspoints.CreateGpsPoint{
				GpsID:     gps.ID,
				Latitude:  node.Lat,
				Longitude: node.Lng,
				Bearing:   computeBearing(prevNode, node),
				Timestamp: time.Now().UnixMilli(),
			}, baseURL)
		}
	}
}

func Run(baseURL string, nodes map[int64]models.Node, adj map[int64][]int64, ctx context.Context) {
	var mu sync.Mutex
	running := make(map[int]struct{})

	// first checks the map before spawning so existing vehicles are untouched
	startGps := func(gps internalgps.GpsResponse) {
		mu.Lock()
		if _, ok := running[gps.ID]; ok {
			mu.Unlock()
			return
		}
		running[gps.ID] = struct{}{}
		mu.Unlock()

		go func() {
			DriveVehicle(baseURL, gps, nodes, adj, ctx)
		}()
		log.Printf("simulator: started gps %d (%s)", gps.ID, gps.SN)
	}

	// initial gps devices load
	gpsDevices, err := apis.FetchGpsDevices(baseURL)
	if err != nil {
		log.Fatalf("failed to fetch GPS devices: %v", err)
	}
	for _, gps := range gpsDevices {
		startGps(gps)
	}

	// subscribe to SSE for instant notification of new GPS devices
	go func() {
		for {
			err := apis.SubscribeToNewDevices(ctx, baseURL, startGps)
			if ctx.Err() != nil {
				return
			}
			log.Printf("simulator: SSE disconnected (%v), reconnecting...", err)
			time.Sleep(2 * time.Second)
		}
	}()

	<-ctx.Done()
}
