package sim

import (
	"context"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	gpspoints "github.com/ize-302/beacon/backend/internal/gps-points"
	"github.com/ize-302/beacon/backend/internal/vehicles"
	"github.com/paulmach/osm"
)

// lineGraph builds n nodes in a straight west-to-east line, each connected to
// its neighbours, so any two nodes are reachable.
func lineGraph(n int) *Graph {
	nodes := make(map[int64]osm.Node, n)
	adj := make(map[int64][]int64, n)
	for i := range n {
		id := int64(i + 1)
		nodes[id] = osm.Node{ID: osm.NodeID(id), Lat: 6.5, Lon: 3.3 + float64(i)/1000}
		var nb []int64
		if i > 0 {
			nb = append(nb, id-1)
		}
		if i < n-1 {
			nb = append(nb, id+1)
		}
		adj[id] = nb
	}
	return NewGraph(nodes, adj)
}

type capturingPoster struct {
	mu         sync.Mutex
	points     []gpspoints.CreateGpsPoint
	batchSizes []int
	err        error
	delay      time.Duration
}

func (c *capturingPoster) sendGpsPoints(ctx context.Context, batch []gpspoints.CreateGpsPoint) error {
	if c.delay > 0 {
		select {
		case <-time.After(c.delay):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.err != nil {
		return c.err
	}
	// Copy: the sender reuses the batch's backing array after we return.
	c.points = append(c.points, batch...)
	c.batchSizes = append(c.batchSizes, len(batch))
	return nil
}

func (c *capturingPoster) count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.points)
}

func (c *capturingPoster) batches() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.batchSizes)
}

// The planner must run searches in parallel but never more than its worker
// count at once — that ceiling is what keeps peak memory bounded regardless of
// fleet size.
func TestPlannerBoundsConcurrency(t *testing.T) {
	const workers = 2
	// Big enough that each search lasts long enough to be sampled mid-flight,
	// small enough that the suite stays fast under -race.
	g := lineGraph(3000)
	p := newPlanner(g, workers)

	var peak atomic.Int64
	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-done:
				return
			default:
				c := int64(p.inFlight())
				for {
					old := peak.Load()
					if c <= old || peak.CompareAndSwap(old, c) {
						break
					}
				}
				time.Sleep(50 * time.Microsecond)
			}
		}
	}()

	var wg sync.WaitGroup
	for i := range 12 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			p.plan(context.Background(), 1, int64(2800+i))
		}()
	}
	wg.Wait()
	close(done)

	if got := peak.Load(); got > workers {
		t.Fatalf("planner ran %d searches at once, cap is %d", got, workers)
	}
	if got := peak.Load(); got < 2 {
		t.Fatalf("planner never ran searches in parallel (peak %d); the cap is throttling too hard", got)
	}
	if p.inFlight() != 0 {
		t.Fatalf("planner leaked %d slots", p.inFlight())
	}
}

func TestPlannerRespectsCancellation(t *testing.T) {
	g := lineGraph(50)
	p := newPlanner(g, 1)

	// Occupy the only slot.
	p.sem <- struct{}{}
	defer func() { <-p.sem }()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	if path := p.plan(ctx, 1, 40); path != nil {
		t.Fatalf("expected nil path on cancellation, got %d nodes", len(path))
	}
	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Fatalf("plan did not return promptly on cancellation: %v", elapsed)
	}
}

// A vehicle must not stall because the API is slow or unreachable.
func TestSenderNeverBlocksVehicle(t *testing.T) {
	slow := &capturingPoster{delay: time.Hour}
	s := newSender(slow, 4, 20*time.Millisecond, 500)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go s.run(ctx)

	start := time.Now()
	for i := range 1000 {
		s.enqueue(gpspoints.CreateGpsPoint{VehicleID: 1, Timestamp: int64(i)})
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("enqueue blocked: 1000 positions took %v", elapsed)
	}
	if s.dropped.Load() == 0 {
		t.Fatal("expected drops once the queue filled")
	}
}

// A failing API must not kill the sender: it used to panic.
func TestSenderSurvivesPostFailure(t *testing.T) {
	failing := &capturingPoster{err: context.DeadlineExceeded}
	s := newSender(failing, 16, 20*time.Millisecond, 500)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go s.run(ctx)

	for i := range 5 {
		s.enqueue(gpspoints.CreateGpsPoint{VehicleID: 1, Timestamp: int64(i)})
	}

	deadline := time.Now().Add(2 * time.Second)
	for s.failed.Load() < 5 && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if got := s.failed.Load(); got < 5 {
		t.Fatalf("sender stopped after a failure: only %d handled", got)
	}

	// Still alive and draining.
	s.enqueue(gpspoints.CreateGpsPoint{VehicleID: 1, Timestamp: 99})
	time.Sleep(200 * time.Millisecond)
	if s.failed.Load() < 6 {
		t.Fatal("sender did not process work after an error")
	}
}

// Many positions must collapse into few requests — that is the whole point of C.
func TestSenderBatchesByInterval(t *testing.T) {
	capture := &capturingPoster{}
	s := newSender(capture, 4096, 50*time.Millisecond, 500)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go s.run(ctx)

	const n = 300
	for i := range n {
		s.enqueue(gpspoints.CreateGpsPoint{VehicleID: 1, Timestamp: int64(i)})
	}

	deadline := time.Now().Add(3 * time.Second)
	for capture.count() < n && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if capture.count() != n {
		t.Fatalf("delivered %d points, want %d", capture.count(), n)
	}
	// One request per interval, not one per point.
	if b := capture.batches(); b >= n/10 {
		t.Fatalf("%d points went out in %d batches; batching is not collapsing requests", n, b)
	}

	// Order must survive batching.
	capture.mu.Lock()
	defer capture.mu.Unlock()
	for i, p := range capture.points {
		if p.Timestamp != int64(i) {
			t.Fatalf("point %d out of order: timestamp %d", i, p.Timestamp)
		}
	}
}

// A burst larger than BatchSize must flush early rather than wait for the tick.
func TestSenderFlushesOnFullBatch(t *testing.T) {
	capture := &capturingPoster{}
	// Interval long enough that a tick-driven flush cannot explain the result.
	s := newSender(capture, 4096, 30*time.Second, 25)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go s.run(ctx)

	for i := range 100 {
		s.enqueue(gpspoints.CreateGpsPoint{VehicleID: 1, Timestamp: int64(i)})
	}

	deadline := time.Now().Add(3 * time.Second)
	for capture.count() < 100 && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if capture.count() < 100 {
		t.Fatalf("size-triggered flush did not happen: only %d delivered", capture.count())
	}
}

// End to end on a synthetic graph: a vehicle should plan a route and emit
// positions along it.
func TestVehicleEmitsPositions(t *testing.T) {
	g := lineGraph(100)
	capture := &capturingPoster{}

	cfg := Config{
		Graph:       g,
		MinInterval: 5 * time.Millisecond,
		MaxInterval: 5 * time.Millisecond,
		SendQueue:   256,
		Seed:        1,
	}
	cfg.applyDefaults()

	w := &world{
		cfg:     cfg,
		graph:   g,
		planner: newPlanner(g, 2),
		sender:  newSender(capture, cfg.SendQueue, cfg.BatchInterval, cfg.BatchSize),
		running: make(map[int]struct{}),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go w.sender.run(ctx)

	go w.driveVehicle(ctx, vehicles.VehicleResponse{ID: 7, PlateNumber: "TEST-7"})

	deadline := time.Now().Add(5 * time.Second)
	for capture.count() < 10 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if capture.count() < 10 {
		t.Fatalf("vehicle emitted only %d positions", capture.count())
	}

	capture.mu.Lock()
	defer capture.mu.Unlock()
	for i, p := range capture.points {
		if p.VehicleID != 7 {
			t.Fatalf("point %d has VehicleID %d, want 7", i, p.VehicleID)
		}
		if p.Latitude == 0 || p.Longitude == 0 {
			t.Fatalf("point %d has zero coordinates", i)
		}
	}
}

// Route choice must be reproducible for a given seed.
func TestRandomNodeIsSeedReproducible(t *testing.T) {
	g := lineGraph(500)

	pick := func() []int64 {
		rng := rand.New(rand.NewSource(42))
		out := make([]int64, 20)
		for i := range out {
			out[i] = g.RandomNode(rng)
		}
		return out
	}

	a, b := pick(), pick()
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("same seed produced different node at %d: %d vs %d", i, a[i], b[i])
		}
	}
}
