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
// Vehicles used to POST inline, which meant a stalled request froze that vehicle
// for the full HTTP timeout. Now they hand the position over and carry on.
//
// There is deliberately one sending goroutine, not a pool: two would let two
// positions from the same vehicle overtake each other, and out-of-order points
// make markers jump backwards on the map. Batching (step C) replaces this with
// one request per interval, which removes the throughput concern entirely.
type sender struct {
	out     chan gpspoints.CreateGpsPoint
	client  poster
	sent    atomic.Uint64
	dropped atomic.Uint64
}

func newSender(client poster, queue int) *sender {
	if queue <= 0 {
		queue = 1024
	}
	return &sender{
		out:    make(chan gpspoints.CreateGpsPoint, queue),
		client: client,
	}
}

// enqueue never blocks. If the queue is full the position is dropped rather than
// stalling the vehicle: a backed-up queue means the API is unreachable, and a
// stale position is superseded by the next one anyway.
func (s *sender) enqueue(p gpspoints.CreateGpsPoint) {
	select {
	case s.out <- p:
	default:
		s.dropped.Add(1)
	}
}

func (s *sender) run(ctx context.Context) {
	report := time.NewTicker(10 * time.Second)
	defer report.Stop()

	var lastSent, lastDropped uint64

	for {
		select {
		case <-ctx.Done():
			return

		case p := <-s.out:
			if err := s.client.sendGpsPosition(ctx, p); err != nil {
				if ctx.Err() != nil {
					return
				}
				// A failed post is not fatal — the next position supersedes it.
				// This used to be a panic, which turned any API blip into a dead
				// simulator.
				s.dropped.Add(1)
				continue
			}
			s.sent.Add(1)

		case <-report.C:
			sent, dropped := s.sent.Load(), s.dropped.Load()
			if sent == lastSent && dropped == lastDropped {
				continue
			}
			log.Printf("simulator: sent %d points (+%d), dropped %d (+%d), queue %d/%d",
				sent, sent-lastSent, dropped, dropped-lastDropped, len(s.out), cap(s.out))
			lastSent, lastDropped = sent, dropped
		}
	}
}
