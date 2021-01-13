//  Copyright (C) 2020 Maker Ecosystem Growth Holdings, INC.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU Affero General Public License as
//  published by the Free Software Foundation, either version 3 of the
//  License, or (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU Affero General Public License for more details.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with this program.  If not, see <http://www.gnu.org/licenses/>.

package graph

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/hashicorp/go-multierror"
)

type ErrNotEnoughSources struct {
	Given int
	Min   int
}

func (e ErrNotEnoughSources) Error() string {
	return fmt.Sprintf(
		"not enough sources to calculate median, %d given but at least %d required",
		e.Given,
		e.Min,
	)
}

type ErrIncompatiblePairs struct {
	Given    Pair
	Expected Pair
}

func (e ErrIncompatiblePairs) Error() string {
	return fmt.Sprintf(
		"unable to calculate median for different pairs, %s given but %s was expected",
		e.Given,
		e.Expected,
	)
}

// MedianAggregatorNode gets Ticks from all of its children and calculates
// median price.
//
//                           -- [Origin A/B]
//                          /
//  [MedianAggregatorNode] ---- [Origin A/B]       -- ...
//                          \                     /
//                           -- [Aggregator A/B] ---- ...
//                                                \
//                                                 -- ...
//
// All children of this node must return a Tick for the same pair.
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

// Children implements the Node interface.
func (n *MedianAggregatorNode) Children() []Node {
	return n.children
}

// AddChild implements the Parent interface.
func (n *MedianAggregatorNode) AddChild(node Node) {
	n.children = append(n.children, node)
}

func (n *MedianAggregatorNode) Pair() Pair {
	return n.pair
}

func (n *MedianAggregatorNode) Tick() AggregatorTick {
	var ts time.Time
	var prices, bids, asks []float64
	var originTicks []OriginTick
	var aggregatorTicks []AggregatorTick
	var err error

	for i, c := range n.children {
		// There is no need to copy errors from ticks to the MedianAggregatorNode
		// because there may be enough remaining ticks to calculate median tick.

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
			aggregatorTick := typedNode.Tick()
			aggregatorTicks = append(aggregatorTicks, aggregatorTick)
			tick = aggregatorTick.Tick
			if aggregatorTick.Error != nil {
				continue
			}
		}

		if !n.pair.Equal(tick.Pair) {
			err = multierror.Append(
				err,
				ErrIncompatiblePairs{Given: tick.Pair, Expected: n.pair},
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
		if i == 0 || tick.Timestamp.Before(ts) {
			ts = tick.Timestamp
		}
	}

	if len(prices) < n.minSources {
		err = multierror.Append(
			err,
			ErrNotEnoughSources{Given: len(prices), Min: n.minSources},
		)
	}

	return AggregatorTick{
		Tick: Tick{
			Pair:      n.pair,
			Price:     median(prices),
			Bid:       median(bids),
			Ask:       median(asks),
			Volume24h: 0,
			Timestamp: ts,
		},
		OriginTicks:     originTicks,
		AggregatorTicks: aggregatorTicks,
		Parameters:      map[string]string{"method": "median", "min": strconv.Itoa(n.minSources)},
		Error:           err,
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
