package graph

import (
	"errors"
	"fmt"
	"sort"

	"github.com/hashicorp/go-multierror"
)

// MedianAggregatorNode gets Ticks from all of its children and calculates
// median price.
//
//                           -- [Origin A/B]
//                          /
//  [MedianAggregatorNode] ---- [Origin A/B]
//                          \
//                           -- [Aggregator A/B]
//
// All children of this node must return tick for the same pair.
type MedianAggregatorNode struct {
	pair       Pair
	minSources int
	children   []Node
}

func NewMedianAggregatorNode(pair Pair, minSources int) *MedianAggregatorNode {
	return &MedianAggregatorNode{
		pair:       pair,
		minSources: minSources,
	}
}

func (n *MedianAggregatorNode) Children() []Node {
	return n.children
}

func (n *MedianAggregatorNode) AddChild(node Node) {
	n.children = append(n.children, node)
}

func (n *MedianAggregatorNode) Pair() Pair {
	return n.pair
}

func (n *MedianAggregatorNode) Tick() IndirectTick {
	var prices, bids, asks []float64
	var originTicks []OriginTick
	var indirectTicks []IndirectTick
	var err error

	for _, c := range n.children {
		var tick Tick
		switch typedNode := c.(type) {
		case Origin:
			originTick := typedNode.Tick()
			originTicks = append(originTicks, originTick)
			tick = originTick.Tick
			if originTick.Error != nil {
				continue
			}
		case Aggregator:
			indirectTick := typedNode.Tick()
			indirectTicks = append(indirectTicks, indirectTick)
			tick = indirectTick.Tick
			if indirectTick.Error != nil {
				continue
			}
		}

		if !n.pair.Equal(tick.Pair) {
			err = multierror.Append(
				err,
				fmt.Errorf(
					"unable to calculate median for different pairs, %s given but %s was expected",
					tick.Pair,
					n.pair,
				),
			)
			continue
		}

		if tick.Price > 0 {
			prices = append(prices, tick.Price)
		}
		if tick.Bid > 0 {
			bids = append(bids, tick.Bid)
		}
		if tick.Ask > 0 {
			asks = append(asks, tick.Ask)
		}
	}

	if len(prices) < n.minSources {
		err = multierror.Append(
			err,
			errors.New("not enough sources to calculate median"),
		)
	}

	return IndirectTick{
		Tick: Tick{
			Pair:      n.pair,
			Price:     median(prices),
			Bid:       median(bids),
			Ask:       median(asks),
			Volume24h: 0,
		},
		OriginTicks:  originTicks,
		IndirectTick: indirectTicks,
		Error:        err,
	}
}

func median(xs []float64) float64 {
	count := len(xs)
	if count == 0 {
		return 0
	}

	sort.Float64s(xs)
	if count%2 == 0 {
		m := count / 2
		x1 := xs[m-1]
		x2 := xs[m]
		return (x1 + x2) / 2
	}

	return xs[(count-1)/2]
}
