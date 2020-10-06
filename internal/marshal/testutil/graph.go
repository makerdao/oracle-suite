package testutil

import (
	"errors"
	"time"

	"github.com/makerdao/gofer/pkg/graph"
)

func Graph(p graph.Pair) *graph.MedianAggregatorNode {
	root := graph.NewMedianAggregatorNode(p, 1)

	on1 := graph.NewOriginNode(graph.OriginPair{Origin: "a", Pair: p})
	on2 := graph.NewOriginNode(graph.OriginPair{Origin: "b", Pair: p})
	in := graph.NewIndirectAggregatorNode(p)
	mn := graph.NewMedianAggregatorNode(p, 1)

	root.AddChild(on1)
	root.AddChild(in)
	root.AddChild(mn)

	in.AddChild(on1)
	mn.AddChild(on1)
	mn.AddChild(on2)

	on1.Ingest(graph.OriginTick{
		Tick:   graph.Tick{
			Pair:      p,
			Price:     10,
			Bid:       10,
			Ask:       10,
			Volume24h: 10,
			Timestamp: time.Unix(10, 0),
		},
		Origin: "a",
		Error:  nil,
	})

	on2.Ingest(graph.OriginTick{
		Tick:   graph.Tick{
			Pair:      p,
			Price:     20,
			Bid:       20,
			Ask:       20,
			Volume24h: 20,
			Timestamp: time.Unix(20, 0),
		},
		Origin: "b",
		Error:  errors.New("something"),
	})

	return root
}
