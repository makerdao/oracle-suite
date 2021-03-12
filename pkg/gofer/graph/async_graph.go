package graph

import (
	"github.com/makerdao/gofer/pkg/gofer"
	"github.com/makerdao/gofer/pkg/gofer/graph/feeder"
	"github.com/makerdao/gofer/pkg/gofer/graph/nodes"
)

// AsyncGraph implements the gofer.Gofer interface. It works just like Graph
// but allows to update prices asynchronously.
type AsyncGraph struct {
	*Graph
	feeder *feeder.Feeder
}

// NewGraph returns a new AsyncGraph instance.
func NewAsyncGraph(g map[gofer.Pair]nodes.Aggregator, f *feeder.Feeder) *AsyncGraph {
	return &AsyncGraph{
		Graph:  NewGraph(g, nil),
		feeder: f,
	}
}

// Start starts asynchronous price updater.
func (a *AsyncGraph) Start() error {
	ns, _ := a.findNodes()
	return a.feeder.Start(ns...)
}

// Start stops asynchronous price updater.
func (a *AsyncGraph) Stop() error {
	a.feeder.Stop()
	return nil
}
