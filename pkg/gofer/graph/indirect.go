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

	"github.com/hashicorp/go-multierror"
)

type ErrTick struct {
	Pair Pair
	Err  error
}

func (e ErrTick) Error() string {
	return fmt.Sprintf(
		"the tick for the %s pair was returned with the following error: %s",
		e.Pair,
		e.Err.Error(),
	)
}

type ErrResolve struct {
	ExpectedPair Pair
	ResolvedPair Pair
}

func (e ErrResolve) Error() string {
	return fmt.Sprintf(
		"the tick was resolved to the %s pair but the %s pair was expected",
		e.ResolvedPair,
		e.ExpectedPair,
	)
}

type ErrInvalidPrice struct {
	Pair Pair
}

func (e ErrInvalidPrice) Error() string {
	return fmt.Sprintf(
		"the calculated price for the %s pair is zero or less",
		e.Pair,
	)
}

type ErrNoCommonPart struct {
	PairA Pair
	PairB Pair
}

func (e ErrNoCommonPart) Error() string {
	return fmt.Sprintf(
		"unable to calculate cross rate for the %s pair with the %s pair, because they have no common part",
		e.PairA,
		e.PairB,
	)
}

type ErrDivByZero struct {
	PairA Pair
	PairB Pair
}

func (e ErrDivByZero) Error() string {
	return fmt.Sprintf(
		"unable to calculate cross rate for the %s pair with the %s pair, because it requires division by zero",
		e.PairA,
		e.PairB,
	)
}

// IndirectAggregatorNode calculates a tick which is a cross rate between all
// child ticks.
//
//                             -- [Origin A/B]
//                            /
//  [IndirectAggregatorNode] ---- [Origin B/C]       -- ...
//                            \                     /
//                             -- [Aggregator C/D] ---- ...
//                                                  \
//                                                   -- ...
//
// For above node, cross rate for the A/D pair will be calculated. It is important
// to add child nodes in the correct order, because ticks will be calculated from
// first to last.
type IndirectAggregatorNode struct {
	pair     Pair
	children []Node
}

func NewIndirectAggregatorNode(pair Pair) *IndirectAggregatorNode {
	return &IndirectAggregatorNode{
		pair: pair,
	}
}

// Children implements the Node interface.
func (n *IndirectAggregatorNode) Children() []Node {
	return n.children
}

// AddChild implements the Parent interface.
func (n *IndirectAggregatorNode) AddChild(node Node) {
	n.children = append(n.children, node)
}

func (n *IndirectAggregatorNode) Pair() Pair {
	return n.pair
}

func (n *IndirectAggregatorNode) Tick() AggregatorTick {
	var ticks []Tick
	var originTicks []OriginTick
	var aggregatorTicks []AggregatorTick
	var err error

	for _, c := range n.children {
		// It's important to copy errors from ticks to the IndirectAggregatorNode,
		// because all of these ticks are required to calculate indirect tick.
		// If there is a problem with any of them, calculated tick won't be
		// reliable.

		switch typedNode := c.(type) {
		case Origin:
			tick := typedNode.Tick()
			originTicks = append(originTicks, tick)
			ticks = append(ticks, tick.Tick)
			if tick.Error != nil {
				err = multierror.Append(
					err,
					ErrTick{
						Pair: tick.Pair,
						Err:  tick.Error,
					},
				)
			}
		case Aggregator:
			tick := typedNode.Tick()
			aggregatorTicks = append(aggregatorTicks, tick)
			ticks = append(ticks, tick.Tick)
			if tick.Error != nil {
				err = multierror.Append(
					err,
					ErrTick{
						Pair: tick.Pair,
						Err:  tick.Error,
					},
				)
			}
		}
	}

	indirectTick, e := crossRate(ticks)
	if e != nil {
		err = multierror.Append(err, e)
	}

	if !indirectTick.Pair.Equal(n.pair) {
		err = multierror.Append(
			err,
			ErrResolve{
				ExpectedPair: n.pair,
				ResolvedPair: indirectTick.Pair,
			},
		)
	}

	if indirectTick.Price <= 0 {
		err = multierror.Append(
			err,
			ErrInvalidPrice{
				Pair: indirectTick.Pair,
			},
		)
	}

	return AggregatorTick{
		Tick:            indirectTick,
		OriginTicks:     originTicks,
		AggregatorTicks: aggregatorTicks,
		Parameters:      map[string]string{"method": "indirect"},
		Error:           err,
	}
}

// crossRate returns calculated tick from the list of ticks. Ticks order is
// important because ticks are calculated from first to last.
//
// TODO: Decide what to do with division by zero during calculating Bid/Ask prices.
//nolint:gocyclo,funlen
func crossRate(t []Tick) (Tick, error) {
	var err error

	if len(t) == 0 {
		return Tick{}, nil
	}

	for i := 0; i < len(t)-1; i++ {
		a := t[i]
		b := t[i+1]

		var pair Pair
		var price, bid, ask float64
		switch {
		case a.Pair.Quote == b.Pair.Quote: // A/C, B/C
			pair.Base = a.Pair.Base
			pair.Quote = b.Pair.Base

			if b.Price > 0 {
				price = a.Price / b.Price
			} else {
				err = multierror.Append(err, ErrDivByZero{a.Pair, b.Pair})
				price = 0
			}

			if b.Bid > 0 {
				bid = a.Bid / b.Bid
			} else {
				bid = 0
			}

			if b.Ask > 0 {
				ask = a.Ask / b.Ask
			} else {
				ask = 0
			}
		case a.Pair.Base == b.Pair.Base: // C/A, C/B
			pair.Base = a.Pair.Quote
			pair.Quote = b.Pair.Quote

			if a.Price > 0 {
				price = b.Price / a.Price
			} else {
				err = multierror.Append(err, ErrDivByZero{a.Pair, b.Pair})
				price = 0
			}

			if a.Bid > 0 {
				bid = b.Bid / a.Bid
			} else {
				bid = 0
			}

			if a.Ask > 0 {
				ask = b.Ask / a.Ask
			} else {
				ask = 0
			}
		case a.Pair.Quote == b.Pair.Base: // A/C, C/B
			pair.Base = a.Pair.Base
			pair.Quote = b.Pair.Quote
			price = a.Price * b.Price
			bid = a.Bid * b.Bid
			ask = a.Ask * b.Ask
		case a.Pair.Base == b.Pair.Quote: // C/A, B/C -> A/B
			pair.Base = a.Pair.Quote
			pair.Quote = b.Pair.Base

			if a.Price > 0 && b.Price > 0 {
				price = (float64(1) / b.Price) / a.Price
			} else {
				err = multierror.Append(err, ErrDivByZero{a.Pair, b.Pair})
				price = 0
			}

			if a.Bid > 0 && b.Bid > 0 {
				bid = (float64(1) / b.Bid) / a.Bid
			} else {
				bid = 0
			}

			if a.Ask > 0 && b.Ask > 0 {
				ask = (float64(1) / b.Ask) / a.Ask
			} else {
				ask = 0
			}
		default:
			err = multierror.Append(err, ErrNoCommonPart{a.Pair, b.Pair})

			return a, err
		}

		b.Pair = pair
		b.Price = price
		b.Bid = bid
		b.Ask = ask
		b.Volume24h = 0
		if a.Timestamp.Before(b.Timestamp) {
			b.Timestamp = a.Timestamp
		}

		t[i+1] = b
	}

	return t[len(t)-1], err
}
