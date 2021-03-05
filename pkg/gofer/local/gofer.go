package local

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/makerdao/gofer/pkg/gofer"
	"github.com/makerdao/gofer/pkg/gofer/local/feeder"
	"github.com/makerdao/gofer/pkg/gofer/local/graph"
)

type ErrPairNotFound struct {
	Pair gofer.Pair
}

func (e ErrPairNotFound) Error() string {
	return fmt.Sprintf("unable to find the %s pair", e.Pair)
}

// Gofer implements the gofer.Gofer interface. It uses a graph structure to
// calculate pairs prices.
type Gofer struct {
	graphs map[gofer.Pair]graph.Aggregator
	feeder *feeder.Feeder
}

// NewGofer returns a new Gofer instance. If the feeder is not nil, then prices
// will be automatically updated when the Price or Prices methods are called.
func NewGofer(g map[gofer.Pair]graph.Aggregator, f *feeder.Feeder) *Gofer {
	return &Gofer{graphs: g, feeder: f}
}

// Nodes implements the gofer.Gofer interface.
func (g *Gofer) Nodes(pairs ...gofer.Pair) (map[gofer.Pair]*gofer.Node, error) {
	nodes, err := g.findNodes(pairs...)
	if err != nil {
		return nil, err
	}

	r := make(map[gofer.Pair]*gofer.Node)
	for _, n := range nodes {
		if n, ok := n.(graph.Aggregator); ok {
			r[n.Pair()] = mapGraphNodes(n)
		}
	}

	return r, nil
}

// Price implements the gofer.Gofer interface.
func (g *Gofer) Tick(pair gofer.Pair) (*gofer.Tick, error) {
	node, ok := g.graphs[pair]
	if !ok {
		return nil, ErrPairNotFound{Pair: pair}
	}

	if g.feeder != nil {
		g.feeder.Feed(node)
	}

	return mapGraphTick(node.Tick()), nil
}

// Prices implements the gofer.Gofer interface.
func (g *Gofer) Ticks(pairs ...gofer.Pair) (map[gofer.Pair]*gofer.Tick, error) {
	nodes, err := g.findNodes(pairs...)
	if err != nil {
		return nil, err
	}

	if g.feeder != nil {
		g.feeder.Feed(nodes...)
	}

	r := make(map[gofer.Pair]*gofer.Tick)
	for _, n := range nodes {
		if n, ok := n.(graph.Aggregator); ok {
			r[n.Pair()] = mapGraphTick(n.Tick())
		}
	}

	return r, nil
}

// Pairs implements the gofer.Gofer interface.
func (g *Gofer) Pairs() ([]gofer.Pair, error) {
	var ps []gofer.Pair
	for p := range g.graphs {
		ps = append(ps, p)
	}
	return ps, nil
}

func (g *Gofer) findNodes(pairs ...gofer.Pair) ([]graph.Node, error) {
	var nodes []graph.Node
	// Return all nodes if no pairs are specified:
	if len(pairs) == 0 {
		for _, node := range g.graphs {
			nodes = append(nodes, node)
		}
		return nodes, nil
	}
	// Find nodes for given pair names:
	for _, pair := range pairs {
		node, ok := g.graphs[pair]
		if !ok {
			return nil, ErrPairNotFound{Pair: pair}
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}

func mapGraphNodes(n graph.Node) *gofer.Node {
	gn := &gofer.Node{
		Type:       strings.TrimLeft(reflect.TypeOf(n).String(), "*"),
		Parameters: make(map[string]string),
	}

	switch typedNode := n.(type) {
	case *graph.IndirectAggregatorNode:
		gn.Type = "indirect"
		gn.Pair = typedNode.Pair()
	case *graph.MedianAggregatorNode:
		gn.Type = "median"
		gn.Pair = typedNode.Pair()
	case *graph.OriginNode:
		gn.Type = "origin"
		gn.Pair = typedNode.OriginPair().Pair
		gn.Parameters["origin"] = typedNode.OriginPair().Origin
	}

	for _, cn := range n.Children() {
		gn.Children = append(gn.Children, mapGraphNodes(cn))
	}

	return gn
}

func mapGraphTick(t interface{}) *gofer.Tick {
	gt := &gofer.Tick{
		Parameters: make(map[string]string),
	}

	switch typedTick := t.(type) {
	case graph.AggregatorTick:
		gt.Type = "aggregator"
		gt.Pair = typedTick.Pair
		gt.Price = typedTick.Price
		gt.Bid = typedTick.Bid
		gt.Ask = typedTick.Ask
		gt.Volume24h = typedTick.Volume24h
		gt.Time = typedTick.Time
		if typedTick.Error != nil {
			gt.Error = typedTick.Error.Error()
		}
		gt.Parameters = typedTick.Parameters
		for _, ct := range typedTick.OriginTicks {
			gt.Ticks = append(gt.Ticks, mapGraphTick(ct))
		}
		for _, ct := range typedTick.AggregatorTicks {
			gt.Ticks = append(gt.Ticks, mapGraphTick(ct))
		}
	case graph.OriginTick:
		gt.Type = "origin"
		gt.Pair = typedTick.Pair
		gt.Price = typedTick.Price
		gt.Bid = typedTick.Bid
		gt.Ask = typedTick.Ask
		gt.Volume24h = typedTick.Volume24h
		gt.Time = typedTick.Time
		if typedTick.Error != nil {
			gt.Error = typedTick.Error.Error()
		}
		gt.Parameters["origin"] = typedTick.Origin
	}

	return gt
}
