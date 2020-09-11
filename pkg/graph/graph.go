package graph

import (
	"sync"
)

// Node represents generics node in a graph.
type Node interface {
	Children() []Node
}

// Parent represents a node to which you can add a child node.
type Parent interface {
	Node
	AddChild(node Node)
}

// Aggregator represents a node which can aggregate ticks from its children.
type Aggregator interface {
	Node
	Pair() Pair
	Tick() IndirectTick
}

// Exchange represents a node which provides tick directly from an exchange.
type Exchange interface {
	Node
	ExchangePair() ExchangePair
	Tick() ExchangeTick
}

func Walk(node Node, fn func(Node)) {
	nodes := map[Node]struct{}{}

	var recur func(Node)
	recur = func(node Node) {
		nodes[node] = struct{}{}
		for _, n := range node.Children() {
			recur(n)
		}
	}
	recur(node)

	for n, _ := range nodes {
		fn(n)
	}
}

func AsyncWalk(node Node, fn func(Node)) {
	wg := sync.WaitGroup{}

	Walk(node, func(node Node) {
		wg.Add(1)
		go func() {
			fn(node)
			wg.Done()
		}()
	})

	wg.Wait()
}
