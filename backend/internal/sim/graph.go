package sim

import (
	"os"

	"github.com/ize-302/osmgraph/osmgraph"
	"github.com/paulmach/osm"
)

// LoadGraph parses an OSM PBF extract into the node/adjacency pair the engine
// drives vehicles over.
//
// this uses an osm library by yours truely: https://github.com/ize-302/osmgraph
func LoadGraph(path string) (map[int64]osm.Node, map[int64][]int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	return osmgraph.GraphBuilder(f, osmgraph.DefaultRoadFilter, osmgraph.DefaultOneway)
}
