package graph

import (
	"fmt"
)

// IndirectAggregatorNode merges Ticks for different pairs and returns one,
// merged pair.
//
//                             -- [Exchange A/B]
//                            /
//  [IndirectAggregatorNode] ---- [Exchange B/C]
//                            \
//                             -- [Aggregator C/D]
//
// For above node, price for pair A/D will be calculated.
type IndirectAggregatorNode struct {
	children []Node
}

func NewIndirectAggregatorNode() *IndirectAggregatorNode {
	return &IndirectAggregatorNode{}
}

func (n *IndirectAggregatorNode) Children() []Node {
	return n.children
}

func (n *IndirectAggregatorNode) AddChild(node Node) {
	n.children = append(n.children, node)
}

func (n *IndirectAggregatorNode) Tick() IndirectTick {
	var ticks []Tick
	var exchangeTicks []ExchangeTick
	var indirectTicks []IndirectTick
	var err error

	for _, c := range n.children {
		switch typedNode := c.(type) {
		case Exchange:
			exchangeTicks = append(exchangeTicks, typedNode.Tick())
			if typedNode.Tick().Error == nil {
				ticks = append(ticks, typedNode.Tick().Tick)
			}
		case Aggregator:
			indirectTicks = append(indirectTicks, typedNode.Tick())
			if typedNode.Tick().Error == nil {
				ticks = append(ticks, typedNode.Tick().Tick)
			}
		}
	}

	indirectTick, err := calcIndirectTick(ticks)
	if err == nil && indirectTick.Price <= 0 {
		err = fmt.Errorf("calculated price for %s is zero or lower", indirectTick.Pair)
	}

	return IndirectTick{
		Tick:          indirectTick,
		ExchangeTicks: exchangeTicks,
		IndirectTick:  indirectTicks,
		Error:         err,
	}
}

func calcIndirectTick(t []Tick) (Tick, error) {
	if len(t) == 0 {
		return Tick{}, nil
	}

	for i := 0; i < len(t)-1; i++ {
		a := t[i]
		b := t[i+1]

		var pair Pair
		var price, bid, ask float64

		switch true {
		case a.Pair.Quote == b.Pair.Quote: // A/C, B/C
			pair.Base = a.Pair.Base
			pair.Quote = b.Pair.Base

			if b.Price > 0 {
				price = a.Price / b.Price
			} else {
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
		case a.Pair.Base == b.Pair.Quote: // C/A, B/C
			pair.Base = a.Pair.Quote
			pair.Quote = b.Pair.Base

			if b.Price > 0 {
				price = (float64(1) / b.Price) / a.Price
			} else {
				price = 0
			}

			if b.Bid > 0 {
				bid = (float64(1) / b.Bid) / a.Bid
			} else {
				bid = 0
			}

			if b.Ask > 0 {
				ask = (float64(1) / b.Ask) / a.Ask
			} else {
				ask = 0
			}
		default:
			return a, fmt.Errorf("unable to merge %s and %s pairs, becuase they don't have any common part", a.Pair, b.Pair)
		}

		b.Pair = pair
		b.Price = price
		b.Bid = bid
		b.Ask = ask
		b.Volume24h = 0

		t[i+1] = b
	}

	return t[len(t)-1], nil
}
