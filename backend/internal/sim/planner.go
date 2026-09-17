package sim

import (
	"context"
	"runtime"
)

// planner bounds how many route searches may run at once.
// A vehicle waiting for a route has nothing else to do, so blocking here costs
// nothing but a slightly longer pause before it sets off again.
type planner struct {
	sem   chan struct{}
	graph *Graph
}

func newPlanner(g *Graph, workers int) *planner {
	if workers <= 0 {
		workers = runtime.NumCPU()
	}
	return &planner{
		sem:   make(chan struct{}, workers),
		graph: g,
	}
}

// plan waits for a free slot and then computes a route. It returns nil if no
// route exists or ctx is cancelled while queueing.
func (p *planner) plan(ctx context.Context, start, goal int64) []int64 {
	select {
	case p.sem <- struct{}{}:
	case <-ctx.Done():
		return nil
	}
	defer func() { <-p.sem }()

	return aStarPath(p.graph, start, goal)
}

// inFlight reports how many searches are running. Used by tests.
func (p *planner) inFlight() int { return len(p.sem) }
