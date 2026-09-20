package sim

import (
	"math"
	"math/rand"
	"os"

	"github.com/ize-302/osmgraph/osmgraph"
	"github.com/paulmach/osm"
)

// Graph is the road network vehicles drive over. It is built once at startup and
// never mutated afterwards, so vehicle goroutines and planners read it
// concurrently without locking.
type Graph struct {
	Nodes map[int64]osm.Node
	Adj   map[int64][]int64

	// keys is the adjacency keyset, materialised once. Rebuilding it per call
	// cost ~14ms against the Lagos extract, and it is needed on every re-plan.
	keys []int64

	// grid answers NearestNode without scanning every node.
	grid nodeGrid
}

// LoadGraph parses an OSM PBF extract into the graph vehicles drive over.
//
// this uses an osm library by yours truely: https://github.com/ize-302/osmgraph
func LoadGraph(path string) (*Graph, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	nodes, adj, err := osmgraph.GraphBuilder(f, osmgraph.DefaultRoadFilter, osmgraph.DefaultOneway)
	if err != nil {
		return nil, err
	}
	return NewGraph(nodes, adj), nil
}

func NewGraph(nodes map[int64]osm.Node, adj map[int64][]int64) *Graph {
	keys := make([]int64, 0, len(adj))
	for k := range adj {
		keys = append(keys, k)
	}
	return &Graph{Nodes: nodes, Adj: adj, keys: keys, grid: buildGrid(nodes, keys)}
}

// RandomNode picks a routable node. The caller passes its own RNG so each
// vehicle's choices are reproducible without sharing a locked global source.
func (g *Graph) RandomNode(rng *rand.Rand) int64 {
	return g.keys[rng.Intn(len(g.keys))]
}

func (g *Graph) Size() int { return len(g.keys) }

// gridCell is the grid's cell size in degrees, roughly 550m. The Lagos extract
// works out to about 150 routable nodes per cell.
const gridCell = 0.005

type cellKey struct{ row, col int32 }

// nodeGrid buckets routable nodes by coordinate so a nearest-node lookup only
// touches the cells around the query instead of the whole graph.
type nodeGrid struct {
	cells                          map[cellKey][]int64
	minRow, maxRow, minCol, maxCol int32
}

func cellOf(lat, lng float64) cellKey {
	return cellKey{int32(math.Floor(lat / gridCell)), int32(math.Floor(lng / gridCell))}
}

func buildGrid(nodes map[int64]osm.Node, ids []int64) nodeGrid {
	g := nodeGrid{cells: make(map[cellKey][]int64)}
	first := true
	for _, id := range ids {
		n, ok := nodes[id]
		if !ok {
			continue
		}
		k := cellOf(n.Lat, n.Lon)
		g.cells[k] = append(g.cells[k], id)
		if first {
			g.minRow, g.maxRow, g.minCol, g.maxCol = k.row, k.row, k.col, k.col
			first = false
			continue
		}
		g.minRow, g.maxRow = min(g.minRow, k.row), max(g.maxRow, k.row)
		g.minCol, g.maxCol = min(g.minCol, k.col), max(g.maxCol, k.col)
	}
	return g
}

// NearestNode returns the routable node closest to the point, or 0 if the graph
// has none. It only considers routable nodes: a vehicle placed on a node with no
// roads attached could never plan a route.
//
// It scans outward one ring of cells at a time and stops as soon as no unseen
// cell can hold anything closer than the best node found.
func (g *Graph) NearestNode(lat, lng float64) int64 {
	gr := &g.grid
	if len(gr.cells) == 0 {
		return 0
	}

	origin := cellOf(lat, lng)
	maxRing := max(
		abs32(origin.row-gr.minRow), abs32(origin.row-gr.maxRow),
		abs32(origin.col-gr.minCol), abs32(origin.col-gr.maxCol),
	)

	var closest int64
	bestD := math.MaxFloat64

	scan := func(k cellKey) {
		for _, id := range gr.cells[k] {
			n := g.Nodes[id]
			dLat, dLng := n.Lat-lat, n.Lon-lng
			if d := dLat*dLat + dLng*dLng; d < bestD {
				bestD, closest = d, id
			}
		}
	}

	// Rings closer than the grid's bounding box hold no cells, so a query outside
	// the mapped area skips straight to the first ring that can.
	r := max(gapTo(origin.row, gr.minRow, gr.maxRow), gapTo(origin.col, gr.minCol, gr.maxCol))

	for ; r <= maxRing; r++ {
		// Anything in ring r is at least r-1 whole cells from the query.
		if r > 1 {
			if lim := float64(r-1) * gridCell; bestD <= lim*lim {
				break
			}
		}
		if r == 0 {
			scan(origin)
			continue
		}

		// Only cells inside the grid's extent can exist, so clip the ring to it.
		colLo := max(origin.col-r, gr.minCol)
		colHi := min(origin.col+r, gr.maxCol)
		for row := max(origin.row-r, gr.minRow); row <= min(origin.row+r, gr.maxRow); row++ {
			if row == origin.row-r || row == origin.row+r {
				for col := colLo; col <= colHi; col++ {
					scan(cellKey{row, col})
				}
				continue
			}
			if c := origin.col - r; c >= gr.minCol && c <= gr.maxCol {
				scan(cellKey{row, c})
			}
			if c := origin.col + r; c >= gr.minCol && c <= gr.maxCol {
				scan(cellKey{row, c})
			}
		}
	}
	return closest
}

// gapTo is how many cells v lies outside [lo, hi], or 0 if it is inside.
func gapTo(v, lo, hi int32) int32 {
	switch {
	case v < lo:
		return lo - v
	case v > hi:
		return v - hi
	}
	return 0
}

func abs32(v int32) int32 {
	if v < 0 {
		return -v
	}
	return v
}
