package graph

import (
	"sync"
)

// Node represents generics node in a graph.
type Node interface {
	Children() []Node
}

// Aggregator represents a node which can aggregate ticks from its children.
type Aggregator interface {
	Node
	Tick() IndirectTick
}

// Exchange represents a node which provides tick directly from an exchange.
type Exchange interface {
	Node
	Tick() ExchangeTick
}

func Walk(node Node, fn func(Node)) {
	for _, n := range node.Children() {
		Walk(n, fn)
	}

	fn(node)
}

func AsyncWalk(node Node, fn func(Node)) {
	wg := sync.WaitGroup{}

	nodes := map[Node]struct{}{}
	Walk(node, func(node Node) {
		nodes[node] = struct{}{}
	})

	for n, _ := range nodes {
		n := n
		wg.Add(1)
		go func() {
			fn(n)
			wg.Done()
		}()
	}

	fn(node)
	wg.Wait()
}
