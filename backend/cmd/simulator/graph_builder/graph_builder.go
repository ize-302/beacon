// Package graphbuilder
package graphbuilder

import (
	"context"
	"log"
	"os"
	"runtime"

	"github.com/ize-302/beacon/backend/cmd/simulator/models"
	"github.com/paulmach/osm"
	"github.com/paulmach/osm/osmpbf"
)

func BuildGraph() (map[int64]models.Node, map[int64][]int64) {
	f, err := os.Open("cmd/simulator/map_data/lagos.osm.pbf")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	nodes := make(map[int64]models.Node)
	adj := make(map[int64][]int64)

	scanner := osmpbf.New(context.Background(), f, runtime.GOMAXPROCS(-1))
	defer scanner.Close()

	for scanner.Scan() {
		switch o := scanner.Object().(type) {
		case *osm.Node:
			nodes[int64(o.ID)] = models.Node{
				ID:  int64(o.ID),
				Lat: o.Lat,
				Lng: o.Lon,
			}
		case *osm.Way:
			tags := o.Tags.Map()
			h, ok := tags["highway"]
			if !ok {
				continue
			}

			switch h {
			case "motorway", "trunk", "primary", "secondary", "tertiary", "residential", "service":
			default:
				continue
			}

			if len(o.Nodes) < 2 {
				continue
			}

			oneway := tags["oneway"]
			// motorways are implicitly one-way in OSM
			isOneway := oneway == "yes" || oneway == "1" || oneway == "true" || h == "motorway"
			isReversed := oneway == "-1"

			for i := 0; i < len(o.Nodes)-1; i++ {
				from := int64(o.Nodes[i].ID)
				to := int64(o.Nodes[i+1].ID)

				if isReversed {
					adj[to] = append(adj[to], from)
				} else {
					adj[from] = append(adj[from], to)
					if !isOneway {
						adj[to] = append(adj[to], from)
					}
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}

	return nodes, adj
}
