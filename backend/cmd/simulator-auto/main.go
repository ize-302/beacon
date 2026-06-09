package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/paulmach/osm"
	"github.com/paulmach/osm/osmpbf"
)

type Node struct {
	ID  int64
	Lat float64
	Lng float64
}

func buildGraph() (map[int64]Node, map[int64][]int64) {
	f, err := os.Open("lagos.osm.pbf")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	nodes := make(map[int64]Node)
	adj := make(map[int64][]int64)

	scanner := osmpbf.New(context.Background(), f, runtime.GOMAXPROCS(-1))
	defer scanner.Close()

	for scanner.Scan() {
		switch o := scanner.Object().(type) {
		case *osm.Node:
			nodes[int64(o.ID)] = Node{
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

			for i := 0; i < len(o.Nodes)-1; i++ {
				from := int64(o.Nodes[i].ID)
				to := int64(o.Nodes[i+1].ID)

				adj[from] = append(adj[from], to)
				adj[to] = append(adj[to], from)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}

	return nodes, adj
}

func pickRandomNode(adj map[int64][]int64) int64 {
	keys := make([]int64, 0, len(adj))
	for k := range adj {
		keys = append(keys, k)
	}
	return keys[rand.Intn(len(keys))]
}

func startSimulation(i int, nodes map[int64]Node, adj map[int64][]int64) {
	current := pickRandomNode(adj)

	for {
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

		fmt.Printf("gps %d Lat: %.4f, Lng: %.4f\n", i+1, node.Lat, node.Lng)

		time.Sleep(time.Second)
	}
}

type Gps struct {
	ID int
}

func main() {
	gpss := []Gps{}
	gpss = append(gpss, Gps{ID: 1})
	gpss = append(gpss, Gps{ID: 2})

	var wg sync.WaitGroup

	for i := range len(gpss) {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			nodes, adj := buildGraph()

			fmt.Printf("Gps %d nodes: %d, adj nodes: %d\n", i+1, len(nodes), len(adj))

			startSimulation(i, nodes, adj)
		}(i)
	}
	wg.Wait()
}
