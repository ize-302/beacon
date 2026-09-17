package sim

import (
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
	return &Graph{Nodes: nodes, Adj: adj, keys: keys}
}

// RandomNode picks a routable node. The caller passes its own RNG so each
// vehicle's choices are reproducible without sharing a locked global source.
func (g *Graph) RandomNode(rng *rand.Rand) int64 {
	return g.keys[rng.Intn(len(g.keys))]
}

func (g *Graph) Size() int { return len(g.keys) }
