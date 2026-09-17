package sim

import (
	"container/heap"
	"math"

	"github.com/paulmach/osm"
)

// aStarPath finds the shortest route by real distance
// A* instead ranks candidate nodes by cost-so-far plus a straight-line (haversine)
// estimate of the remaining distance to the goal, so the search stays biased
// toward the goal. Straight-line distance never
// overestimates the real road distance, which keeps the result the true
// shortest path rather than an approximation.
//
// Returns a path from start to goal inclusive, or nil if none exists.
func aStarPath(g *Graph, start, goal int64) []int64 {
	if start == goal {
		return []int64{start}
	}
	goalNode, ok := g.Nodes[goal]
	if !ok {
		return nil
	}

	open := &nodeQueue{{node: start, f: 0}}
	heap.Init(open)

	cameFrom := make(map[int64]int64)
	gScore := map[int64]float64{start: 0}
	closed := make(map[int64]bool)

	for open.Len() > 0 {
		cur := heap.Pop(open).(queueItem).node
		if cur == goal {
			return reconstructPath(cameFrom, start, goal)
		}
		if closed[cur] {
			continue // a cheaper entry for this node was already processed
		}
		closed[cur] = true
		curNode := g.Nodes[cur]

		for _, nb := range g.Adj[cur] {
			if closed[nb] {
				continue
			}
			nbNode, ok := g.Nodes[nb]
			if !ok {
				continue
			}

			tentative := gScore[cur] + haversineDistance(curNode, nbNode)
			if existing, seen := gScore[nb]; seen && tentative >= existing {
				continue
			}
			gScore[nb] = tentative
			cameFrom[nb] = cur
			heap.Push(open, queueItem{node: nb, f: tentative + haversineDistance(nbNode, goalNode)})
		}
	}
	return nil // no path exists
}

func reconstructPath(cameFrom map[int64]int64, start, goal int64) []int64 {
	path := []int64{goal}
	for cur := goal; cur != start; {
		prev, ok := cameFrom[cur]
		if !ok {
			return nil
		}
		path = append([]int64{prev}, path...)
		cur = prev
	}
	return path
}

// queueItem is a candidate node in the open set, ranked by f = cost-so-far +
// heuristic. Nodes can be pushed more than once as cheaper routes to them are
// found; the stale, more expensive copies are skipped via the closed set
// rather than removed from the heap, which is simpler than a heap that
// supports decrease-key.
type queueItem struct {
	node int64
	f    float64
}

type nodeQueue []queueItem

func (q nodeQueue) Len() int           { return len(q) }
func (q nodeQueue) Less(i, j int) bool { return q[i].f < q[j].f }
func (q nodeQueue) Swap(i, j int)      { q[i], q[j] = q[j], q[i] }
func (q *nodeQueue) Push(x any)        { *q = append(*q, x.(queueItem)) }
func (q *nodeQueue) Pop() any {
	old := *q
	n := len(old)
	item := old[n-1]
	*q = old[:n-1]
	return item
}

// haversineDistance is great-circle distance in meters. Used both as edge
// cost and as the A* heuristic (straight-line distance to goal), since it is
// cheap and never overestimates the real road distance between two points.
//
// a = sin²(Δlat / 2) + cos(lat1) × cos(lat2) × sin²(Δlon / 2)
// c = 2 × atan2(√a, √(1 − a))
// distance = R × c
func haversineDistance(a, b osm.Node) float64 {
	const earthRadiusM = 6371000.0
	lat1 := a.Lat * math.Pi / 180
	lat2 := b.Lat * math.Pi / 180
	dLat := (b.Lat - a.Lat) * math.Pi / 180
	dLon := (b.Lon - a.Lon) * math.Pi / 180

	h := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*math.Sin(dLon/2)*math.Sin(dLon/2)
	return earthRadiusM * 2 * math.Atan2(math.Sqrt(h), math.Sqrt(1-h))
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
