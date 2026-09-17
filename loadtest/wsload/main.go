// Command wsload load-tests the /ws position broadcast fan-out.
//
// It opens N concurrent websocket connections against the API, ramping them
// up over a configurable window so the hub sees a realistic connect surge
// rather than a single instant, then holds them open while decoding every
// PositionFrame it receives and measuring end-to-end latency (now minus each
// point's timestamp) and throughput. It reports periodically and once more
// at the end.
//
// The API's Hub drops frames on a full per-client queue rather than blocking
// (see internal/ws), so a struggling server shows up here as rising latency
// and stalled frames/s long before connections actually drop.
//
//	go run ./cmd/wsload -url ws://127.0.0.1:8081/ws -conns 2000 -ramp 30s -duration 5m
package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"os/signal"
	"sort"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
)

// positionFrame mirrors the wire shape of gpspoints.PositionFrame
// (backend/internal/gps-points/dto.go). It's copied rather than imported: the
// backend package lives under internal/, which Go only lets code under the
// same backend/ import root reach, and this module intentionally lives
// outside that so it isn't tied to the app's own build. Keep this in sync by
// hand if that wire format changes.
type positionFrame struct {
	Type   string `json:"type"`
	Points []struct {
		VehicleID int     `json:"vehicle_id"`
		Bearing   float64 `json:"bearing"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		Timestamp int64   `json:"timestamp"`
	} `json:"points"`
}

type stats struct {
	connected    int64
	framesTotal  int64
	pointsTotal  int64
	disconnects  int64
	dialErrors   int64
	decodeErrors int64

	mu     sync.Mutex
	window []int64 // per-point latencies (ms) accumulated since the last report
}

// recordLatency caps the window rather than growing it unbounded under heavy
// traffic; a report only needs enough samples for a reasonable percentile.
func (s *stats) recordLatency(ms int64) {
	const maxSamples = 200_000
	s.mu.Lock()
	if len(s.window) < maxSamples {
		s.window = append(s.window, ms)
	}
	s.mu.Unlock()
}

func (s *stats) drainWindow() []int64 {
	s.mu.Lock()
	w := s.window
	s.window = nil
	s.mu.Unlock()
	sort.Slice(w, func(i, j int) bool { return w[i] < w[j] })
	return w
}

func percentile(sorted []int64, p float64) int64 {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(p * float64(len(sorted)))
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

// connRegistry tracks live connections so the test can force them all closed
// at the end. gorilla's ReadMessage blocks on the network and ignores
// context cancellation, so closing the conn is the only way to unblock a
// worker's read loop on shutdown.
type connRegistry struct {
	mu    sync.Mutex
	conns map[*websocket.Conn]struct{}
}

func newConnRegistry() *connRegistry {
	return &connRegistry{conns: make(map[*websocket.Conn]struct{})}
}

func (r *connRegistry) add(c *websocket.Conn) {
	r.mu.Lock()
	r.conns[c] = struct{}{}
	r.mu.Unlock()
}

func (r *connRegistry) remove(c *websocket.Conn) {
	r.mu.Lock()
	delete(r.conns, c)
	r.mu.Unlock()
}

func (r *connRegistry) closeAll() {
	r.mu.Lock()
	defer r.mu.Unlock()
	for c := range r.conns {
		c.Close()
	}
}

func worker(ctx context.Context, url string, st *stats, reg *connRegistry, reconnect bool) {
	for {
		if ctx.Err() != nil {
			return
		}

		conn, _, err := websocket.DefaultDialer.DialContext(ctx, url, nil)
		if err != nil {
			atomic.AddInt64(&st.dialErrors, 1)
			select {
			case <-ctx.Done():
				return
			case <-time.After(2 * time.Second):
			}
			continue
		}

		reg.add(conn)
		atomic.AddInt64(&st.connected, 1)
		readLoop(conn, st)
		reg.remove(conn)
		conn.Close()
		atomic.AddInt64(&st.connected, -1)
		atomic.AddInt64(&st.disconnects, 1)

		if ctx.Err() != nil || !reconnect {
			return
		}
	}
}

// readLoop returns once the connection errors, whether that's the server
// closing it, a network blip, or the test forcing it shut on shutdown.
func readLoop(conn *websocket.Conn, st *stats) {
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}

		var frame positionFrame
		if err := json.Unmarshal(msg, &frame); err != nil {
			atomic.AddInt64(&st.decodeErrors, 1)
			continue
		}

		now := time.Now().UnixMilli()
		atomic.AddInt64(&st.framesTotal, 1)
		atomic.AddInt64(&st.pointsTotal, int64(len(frame.Points)))
		for _, p := range frame.Points {
			st.recordLatency(now - p.Timestamp)
		}
	}
}

func main() {
	url := flag.String("url", "ws://127.0.0.1:8081/ws", "websocket endpoint to load")
	conns := flag.Int("conns", 500, "number of concurrent connections")
	ramp := flag.Duration("ramp", 30*time.Second, "time to spread connection opens over")
	duration := flag.Duration("duration", 5*time.Minute, "how long to hold connections once ramp completes")
	reportEvery := flag.Duration("report", 5*time.Second, "stats report interval")
	reconnect := flag.Bool("reconnect", false, "reconnect a connection if it drops, instead of leaving it closed")
	flag.Parse()

	rootCtx, stopSignals := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopSignals()
	workCtx, cancelWork := context.WithCancel(rootCtx)
	defer cancelWork()

	st := &stats{}
	reg := newConnRegistry()

	log.Printf("wsload: opening %d connections to %s over %s, holding for %s", *conns, *url, *ramp, *duration)

	var wg sync.WaitGroup
	rampInterval := time.Duration(0)
	if *conns > 0 {
		rampInterval = *ramp / time.Duration(*conns)
	}

ramp:
	for i := 0; i < *conns; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			worker(workCtx, *url, st, reg, *reconnect)
		}()

		if rampInterval <= 0 {
			continue
		}
		select {
		case <-rootCtx.Done():
			break ramp
		case <-time.After(rampInterval):
		}
	}

	log.Printf("wsload: ramp complete, %d connected", atomic.LoadInt64(&st.connected))

	reportTicker := time.NewTicker(*reportEvery)
	defer reportTicker.Stop()
	testDeadline := time.After(*duration)

	var lastFrames, lastPoints int64
	lastReport := time.Now()

	report := func() {
		frames := atomic.LoadInt64(&st.framesTotal)
		points := atomic.LoadInt64(&st.pointsTotal)
		elapsed := time.Since(lastReport).Seconds()
		window := st.drainWindow()

		log.Printf(
			"wsload: connected=%d frames/s=%.1f points/s=%.1f latency_ms(p50=%d p95=%d p99=%d max=%d) disconnects=%d dial_err=%d decode_err=%d",
			atomic.LoadInt64(&st.connected),
			float64(frames-lastFrames)/elapsed,
			float64(points-lastPoints)/elapsed,
			percentile(window, 0.50), percentile(window, 0.95), percentile(window, 0.99),
			last(window),
			atomic.LoadInt64(&st.disconnects), atomic.LoadInt64(&st.dialErrors), atomic.LoadInt64(&st.decodeErrors),
		)
		lastFrames, lastPoints = frames, points
		lastReport = time.Now()
	}

loop:
	for {
		select {
		case <-rootCtx.Done():
			break loop
		case <-testDeadline:
			break loop
		case <-reportTicker.C:
			report()
		}
	}

	log.Printf("wsload: shutting down, closing %d connections", atomic.LoadInt64(&st.connected))
	cancelWork()   // stop reconnect/dial attempts
	reg.closeAll() // unblock every worker's ReadMessage

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		log.Printf("wsload: workers did not exit within 10s, reporting anyway")
	}

	report()
	log.Printf("wsload: done, totals frames=%d points=%d disconnects=%d dial_err=%d decode_err=%d",
		atomic.LoadInt64(&st.framesTotal), atomic.LoadInt64(&st.pointsTotal),
		atomic.LoadInt64(&st.disconnects), atomic.LoadInt64(&st.dialErrors), atomic.LoadInt64(&st.decodeErrors))
}

func last(s []int64) int64 {
	if len(s) == 0 {
		return 0
	}
	return s[len(s)-1]
}
