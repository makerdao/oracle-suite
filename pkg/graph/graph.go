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
	Tick() AggregatorTick
}

// Origin represents a node which provides tick directly from an origin.
type Origin interface {
	Node
	OriginPair() OriginPair
	Tick() OriginTick
}

func Walk(fn func(Node), nodes ...Node) {
	r := map[Node]struct{}{}

	for _, node := range nodes {
		var recur func(Node)
		recur = func(node Node) {
			r[node] = struct{}{}
			for _, n := range node.Children() {
				recur(n)
			}
		}
		recur(node)
	}

	for n := range r {
		fn(n)
	}
}

func AsyncWalk(fn func(Node), node ...Node) {
	wg := sync.WaitGroup{}

	Walk(func(node Node) {
		wg.Add(1)
		go func() {
			fn(node)
			wg.Done()
		}()
	}, node...)

	wg.Wait()
}
