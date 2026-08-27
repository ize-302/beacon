package sim

import (
	"context"
	"log"
	"sync/atomic"
	"time"

	gpspoints "github.com/ize-302/beacon/backend/internal/gps-points"
)

// sender owns the only path from vehicles to the API.
//
// Vehicles hand positions over and carry on; they never block on HTTP. The
// sender accumulates them and posts one batch per interval, so request rate is
// fixed by the clock rather than by fleet size — 500 vehicles emitting a point a
// second is 5 requests/sec at a 200ms interval, not 500.
//
// There is deliberately one sending goroutine, not a pool: two would let two
// batches overtake each other, and out-of-order points make markers jump
// backwards on the map.
type sender struct {
	out      chan gpspoints.CreateGpsPoint
	client   poster
	interval time.Duration
	maxBatch int

	sent    atomic.Uint64 // accepted by the API
	dropped atomic.Uint64 // queue was full; never left the simulator
	failed  atomic.Uint64 // posted but the API rejected or timed out
	batches atomic.Uint64
}

func newSender(client poster, queue int, interval time.Duration, maxBatch int) *sender {
	if queue <= 0 {
		queue = 1024
	}
	if interval <= 0 {
		interval = 200 * time.Millisecond
	}
	if maxBatch <= 0 {
		maxBatch = 500
	}
	return &sender{
		out:      make(chan gpspoints.CreateGpsPoint, queue),
		client:   client,
		interval: interval,
		maxBatch: maxBatch,
	}
}

// enqueue never blocks. If the queue is full the position is dropped rather than
// stalling the vehicle: a backed-up queue means the API cannot keep up, and a
// stale position is superseded by the next one anyway.
func (s *sender) enqueue(p gpspoints.CreateGpsPoint) {
	select {
	case s.out <- p:
	default:
		s.dropped.Add(1)
	}
}

func (s *sender) run(ctx context.Context) {
	flush := time.NewTicker(s.interval)
	defer flush.Stop()
	report := time.NewTicker(10 * time.Second)
	defer report.Stop()

	batch := make([]gpspoints.CreateGpsPoint, 0, s.maxBatch)

	// send posts the pending batch and clears it. The client marshals before
	// returning, so reusing the backing array afterwards is safe.
	send := func() {
		if len(batch) == 0 {
			return
		}
		if err := s.client.sendGpsPoints(ctx, batch); err != nil {
			if ctx.Err() == nil {
				// Not fatal — the next positions supersede these. This used to
				// be a panic, which turned any API blip into a dead simulator.
				s.failed.Add(uint64(len(batch)))
			}
		} else {
			s.sent.Add(uint64(len(batch)))
			s.batches.Add(1)
		}
		batch = batch[:0]
	}

	var lastSent, lastDropped, lastFailed, lastBatches uint64

	for {
		select {
		case <-ctx.Done():
			return

		case p := <-s.out:
			batch = append(batch, p)
			// Flush early on a full batch so a burst does not wait for the tick.
			if len(batch) >= s.maxBatch {
				send()
			}

		case <-flush.C:
			send()

		case <-report.C:
			sent, dropped := s.sent.Load(), s.dropped.Load()
			failed, batches := s.failed.Load(), s.batches.Load()
			if sent == lastSent && dropped == lastDropped && failed == lastFailed {
				continue
			}
			log.Printf("simulator: sent %d points in %d batches (+%d/+%d), dropped %d (+%d), failed %d (+%d), queue %d/%d",
				sent, batches, sent-lastSent, batches-lastBatches,
				dropped, dropped-lastDropped,
				failed, failed-lastFailed,
				len(s.out), cap(s.out))
			lastSent, lastDropped, lastFailed, lastBatches = sent, dropped, failed, batches
		}
	}
}
