package sim

import (
	"math"

	"github.com/paulmach/osm"
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
func computeBearing(from, to osm.Node) float64 {
	lat1 := from.Lat * math.Pi / 180
	lng1 := from.Lon * math.Pi / 180
	lat2 := to.Lat * math.Pi / 180
	lng2 := to.Lon * math.Pi / 180

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

// closestNode is a linear scan over every node in the graph. It runs once per
// vehicle when one is admitted with a known last position, so the cost is paid
// rarely — a spatial index replaces it when vehicles can be spawned at an
// arbitrary point on the map.
func closestNode(nodes map[int64]osm.Node, lat, lng float64) int64 {
	var closest int64
	minDist := math.MaxFloat64
	for id, n := range nodes {
		latDiff := n.Lat - lat
		lngDiff := n.Lon - lng
		d := (latDiff * latDiff) + (lngDiff * lngDiff)
		if d < minDist {
			minDist = d
			closest = id
		}
	}
	return closest
}
