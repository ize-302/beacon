// Package sim drives simulated vehicles along the road graph.
package sim

import (
	"context"
	"log"
	"math/rand"
	"sync"
	"time"

	internalgps "github.com/ize-302/beacon/backend/internal/gps"
	gpspoints "github.com/ize-302/beacon/backend/internal/gps-points"
)

// Config tunes the simulation. Zero values fall back to the defaults below.
type Config struct {
	BaseURL string
	Graph   *Graph

	// Planners caps concurrent route searches. Peak planning memory is roughly
	// Planners * 61MB, so raise it for throughput and lower it for a smaller
	// footprint. Defaults to runtime.NumCPU().
	Planners int

	// SendQueue is how many positions may be buffered for the API before new
	// ones are dropped. Defaults to 1024.
	SendQueue int

	// BatchInterval is how often pending positions are posted. Request rate is
	// fixed by this rather than by fleet size. Defaults to 200ms.
	BatchInterval time.Duration

	// BatchSize forces an early flush once this many positions are pending, so a
	// burst does not wait for the interval. Defaults to 500.
	BatchSize int

	// MinInterval/MaxInterval bound how long a vehicle waits between hops. Each
	// vehicle picks a fixed interval in this range when it is admitted.
	MinInterval time.Duration
	MaxInterval time.Duration

	// Seed makes vehicle route choices reproducible. Defaults to the clock.
	Seed int64
}

func (c *Config) applyDefaults() {
	if c.SendQueue <= 0 {
		c.SendQueue = 1024
	}
	if c.BatchInterval <= 0 {
		c.BatchInterval = 200 * time.Millisecond
	}
	if c.BatchSize <= 0 {
		c.BatchSize = 500
	}
	if c.MinInterval <= 0 {
		c.MinInterval = 1 * time.Second
	}
	if c.MaxInterval < c.MinInterval {
		c.MaxInterval = 9 * time.Second
	}
	if c.Seed == 0 {
		c.Seed = time.Now().UnixNano()
	}
}

type world struct {
	cfg     Config
	graph   *Graph
	planner *planner
	sender  *sender
	client  *apiClient

	mu      sync.Mutex
	running map[int]struct{}
}

// Run drives every registered GPS device until ctx is cancelled.
func Run(ctx context.Context, cfg Config) error {
	cfg.applyDefaults()

	client := newAPIClient(cfg.BaseURL)
	w := &world{
		cfg:     cfg,
		graph:   cfg.Graph,
		planner: newPlanner(cfg.Graph, cfg.Planners),
		sender:  newSender(client, cfg.SendQueue, cfg.BatchInterval, cfg.BatchSize),
		client:  client,
		running: make(map[int]struct{}),
	}

	go w.sender.run(ctx)

	// initial gps devices load
	devices, err := client.fetchGpsDevices()
	if err != nil {
		return err
	}
	for _, gps := range devices {
		w.startGps(ctx, gps)
	}

	// subscribe to SSE for instant notification of new GPS devices
	go func() {
		for {
			err := client.subscribeToNewDevices(ctx, func(gps internalgps.GpsResponse) {
				w.startGps(ctx, gps)
			})
			if ctx.Err() != nil {
				return
			}
			log.Printf("simulator: SSE disconnected (%v), reconnecting...", err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(2 * time.Second):
			}
		}
	}()

	<-ctx.Done()
	return nil
}

// startGps checks the map before spawning so existing vehicles are untouched.
func (w *world) startGps(ctx context.Context, gps internalgps.GpsResponse) {
	w.mu.Lock()
	if _, ok := w.running[gps.ID]; ok {
		w.mu.Unlock()
		return
	}
	w.running[gps.ID] = struct{}{}
	w.mu.Unlock()

	go w.driveVehicle(ctx, gps)
	log.Printf("simulator: started gps %d (%s)", gps.ID, gps.SN)
}

func (w *world) driveVehicle(ctx context.Context, gps internalgps.GpsResponse) {
	// Seeded per vehicle so a given seed reproduces the same route choices, and
	// so vehicles do not contend on the locked global source.
	rng := rand.New(rand.NewSource(w.cfg.Seed + int64(gps.ID)))

	var current int64
	if gps.LastCoordinate != nil {
		current = closestNode(w.graph.Nodes, gps.LastCoordinate.Latitude, gps.LastCoordinate.Longitude)
	} else {
		current = w.graph.RandomNode(rng)
	}

	// int64 throughout: a 8s spread is 8e9 nanoseconds, which overflows a 32-bit
	// int and would make rng.Intn panic on a 32-bit build.
	interval := w.cfg.MinInterval
	if spread := int64(w.cfg.MaxInterval - w.cfg.MinInterval); spread > 0 {
		interval += time.Duration(rng.Int63n(spread + 1))
	}
	t := time.NewTicker(interval)
	defer t.Stop()

	var path []int64

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			// determine new path when current one is exhausted
			for len(path) == 0 {
				if ctx.Err() != nil {
					return
				}
				dest := w.graph.RandomNode(rng)
				if dest == current {
					continue
				}
				// Route searches are capped globally; this may wait for a slot.
				path = w.planner.plan(ctx, current, dest)
				if len(path) > 1 {
					path = path[1:] // drop current node
				} else {
					path = nil
				}
			}

			prevNode := w.graph.Nodes[current]
			current = path[0]
			path = path[1:]

			node, ok := w.graph.Nodes[current]
			if !ok {
				continue
			}
			w.sender.enqueue(gpspoints.CreateGpsPoint{
				GpsID:     gps.ID,
				Latitude:  node.Lat,
				Longitude: node.Lon,
				Bearing:   computeBearing(prevNode, node),
				Timestamp: time.Now().UnixMilli(),
			})
		}
	}
}
