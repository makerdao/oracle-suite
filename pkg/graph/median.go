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
//                           -- [Exchange A/B]
//                          /
//  [MedianAggregatorNode] ---- [Exchange A/B]
//                          \
//                           -- [Aggregator A/B]
//
// All children of this node must return tick for the same pair.
type MedianAggregatorNode struct {
	minSources int
	children   []Node
}

func NewMedianAggregatorNode(minSources int) *MedianAggregatorNode {
	return &MedianAggregatorNode{
		minSources: minSources,
	}
}

func (n *MedianAggregatorNode) Children() []Node {
	return n.children
}

func (n *MedianAggregatorNode) AddChild(node Node) {
	n.children = append(n.children, node)
}

func (n *MedianAggregatorNode) Tick() IndirectTick {
	var pair Pair
	var prices, bids, asks []float64
	var exchangeTicks []ExchangeTick
	var indirectTicks []IndirectTick
	var err error

	for _, c := range n.children {
		var tick Tick
		switch typedNode := c.(type) {
		case Exchange:
			exchangeTick := typedNode.Tick()
			exchangeTicks = append(exchangeTicks, exchangeTick)
			tick = exchangeTick.Tick

			if exchangeTick.Error != nil {
				err = multierror.Append(err, exchangeTick.Error)
				continue
			}
		case Aggregator:
			indirectTick := typedNode.Tick()
			indirectTicks = append(indirectTicks, indirectTick)
			tick = indirectTick.Tick

			if indirectTick.Error != nil {
				err = multierror.Append(err, indirectTick.Error)
				continue
			}
		}

		if pair.Empty() {
			pair = tick.Pair
		} else if !pair.Equal(tick.Pair) {
			err = multierror.Append(
				err,
				fmt.Errorf("unable to calculate median for different pairs, %s and %s given", pair, tick.Pair),
			)
			continue
		}

		prices = append(prices, tick.Price)
		bids = append(bids, tick.Bid)
		asks = append(asks, tick.Ask)
	}

	if len(prices) < n.minSources {
		err = multierror.Append(err, errors.New("not enough sources to calculate median"))
	}

	return IndirectTick{
		Tick: Tick{
			Pair:      pair,
			Price:     median(filterOutZeros(prices)),
			Bid:       median(filterOutZeros(bids)),
			Ask:       median(filterOutZeros(asks)),
			Volume24h: 0,
		},
		ExchangeTicks: exchangeTicks,
		IndirectTick:  indirectTicks,
		Error:         err,
	}
}

func filterOutZeros(xs []float64) []float64 {
	var r []float64
	for _, x := range xs {
		if x > 0 {
			r = append(r, x)
		}
	}

	return r
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
