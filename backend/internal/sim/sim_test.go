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

	// The reason has to be recoverable: the counter alone says nothing about why.
	if msg, _ := s.lastErr.Load().(string); msg != context.DeadlineExceeded.Error() {
		t.Fatalf("lastErr = %q, want %q", msg, context.DeadlineExceeded.Error())
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

// The grid must agree with a brute-force scan, including for queries far outside
// the mapped area where the search has to walk many empty rings.
func TestNearestNodeMatchesBruteForce(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	nodes := make(map[int64]osm.Node)
	adj := make(map[int64][]int64)
	for i := range 3000 {
		id := int64(i + 1)
		nodes[id] = osm.Node{ID: osm.NodeID(id), Lat: 6.3 + rng.Float64()*0.4, Lon: 3.1 + rng.Float64()*0.5}
		adj[id] = []int64{1}
	}
	g := NewGraph(nodes, adj)

	dist := func(id int64, lat, lng float64) float64 {
		n := nodes[id]
		return (n.Lat-lat)*(n.Lat-lat) + (n.Lon-lng)*(n.Lon-lng)
	}

	for range 300 {
		// mostly inside the box, sometimes well outside it
		lat, lng := 6.2+rng.Float64()*0.6, 3.0+rng.Float64()*0.7
		if rng.Intn(5) == 0 {
			lat, lng = rng.Float64()*20-10, rng.Float64()*20-10
		}

		want := int64(0)
		for id := range nodes {
			if want == 0 || dist(id, lat, lng) < dist(want, lat, lng) {
				want = id
			}
		}
		if got := g.NearestNode(lat, lng); dist(got, lat, lng) != dist(want, lat, lng) {
			t.Fatalf("NearestNode(%f, %f) = %d, brute force says %d", lat, lng, got, want)
		}
	}
}

func TestNearestNodeEmptyGraph(t *testing.T) {
	if got := NewGraph(nil, nil).NearestNode(6.5, 3.3); got != 0 {
		t.Fatalf("empty graph returned node %d, want 0", got)
	}
}

// A fleet must be admitted at the configured rate rather than all at once.
func TestAdmitterPacesStarts(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	a := newAdmitter(ctx, 200) // one slot every 5ms

	start := time.Now()
	var wg sync.WaitGroup
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			a.wait(ctx)
		}()
	}
	wg.Wait()

	// 20 slots at 5ms each is ~100ms; without pacing this returns instantly.
	if elapsed := time.Since(start); elapsed < 80*time.Millisecond {
		t.Fatalf("20 vehicles admitted in %v, expected pacing near 100ms", elapsed)
	}
}

func TestAdmitterWaitRespectsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	a := newAdmitter(ctx, 1) // slow enough that the wait must be interrupted

	done := make(chan bool)
	go func() { done <- a.wait(ctx) }()

	cancel()
	select {
	case ok := <-done:
		if ok {
			t.Fatal("wait reported a slot after cancellation")
		}
	case <-time.After(time.Second):
		t.Fatal("wait did not return after cancellation")
	}
}

func TestNilAdmitterNeverBlocks(t *testing.T) {
	var a *admitter
	if !a.wait(context.Background()) {
		t.Fatal("nil admitter should admit immediately")
	}
}

// A vehicle must set off as soon as it is admitted, not after its first
// interval. The interval here is far longer than the test waits.
func TestVehicleMovesImmediatelyAfterAdmission(t *testing.T) {
	g := lineGraph(100)
	capture := &capturingPoster{}

	cfg := Config{
		Graph:         g,
		MinInterval:   time.Minute,
		MaxInterval:   time.Minute,
		SendQueue:     16,
		BatchInterval: 10 * time.Millisecond,
		Seed:          1,
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
	go w.driveVehicle(ctx, vehicles.VehicleResponse{ID: 3, PlateNumber: "TEST-3"})

	deadline := time.Now().Add(2 * time.Second)
	for capture.count() < 1 && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if capture.count() < 1 {
		t.Fatal("no position within 2s of admission; vehicle is waiting for its interval")
	}
}
