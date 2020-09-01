package node

import "github.com/makerdao/gofer/pkg/exchange"

type Node interface {
	Children() []Node
	Reduce()
}

type PriceNode interface {
	Node
	Price() float64
}

func Walk(node Node, fn func(Node)) {
	for _, n := range node.Children() {
		Walk(n, fn)
	}

	fn(node)
}

type MedianAggregatorNode struct {
	prices []float64
	children []PriceNode
}

func (m *MedianAggregatorNode) Children() []Node {
	r := make([]Node, len(m.children))
	for i, n := range m.children {
		r[i] = n.(Node)
	}

	return r
}

func (m *MedianAggregatorNode) Reduce() {
	for _, n := range m.children {
		m.prices = append(m.prices, n.Price())
	}

	m.children = []PriceNode{}
}

type ExchangeNode struct {
	handler exchange.Handler
	price   float64
}

func (e ExchangeNode) Children() []Node {
	return []Node{}
}

func (e ExchangeNode) Reduce() {
	// do nothing
}

func (e ExchangeNode) Price() float64 {
	return 10.0
}

