package sim

import (
	"context"
	"runtime"
)

// planner bounds how many route searches may run at once.
//
// A single bfsPath over the Lagos extract allocates ~29MB on average and up to
// ~61MB when it has to explore the whole graph. Vehicle goroutines re-plan
// independently, so without a cap the peak scales with fleet size — 500 vehicles
// re-planning together would need ~30GB. Capping concurrency holds the ceiling at
// roughly workers * 61MB no matter how many vehicles exist.
//
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

	return bfsPath(p.graph.Adj, start, goal)
}

// inFlight reports how many searches are running. Used by tests.
func (p *planner) inFlight() int { return len(p.sem) }
