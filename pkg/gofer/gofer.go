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

package gofer

import (
	"fmt"

	"github.com/makerdao/gofer/pkg/graph"
)

type Gofer struct {
	graphs   map[graph.Pair]graph.Aggregator
	ingestor *graph.Feeder
}

func NewGofer(graphs map[graph.Pair]graph.Aggregator, ingestor *graph.Feeder) *Gofer {
	return &Gofer{
		graphs:   graphs,
		ingestor: ingestor,
	}
}

func (g *Gofer) Graphs() map[graph.Pair]graph.Aggregator {
	return g.graphs
}

func (g *Gofer) Ingestor() *graph.Feeder {
	return g.ingestor
}

func (g *Gofer) Pairs() []graph.Pair {
	var pairs []graph.Pair
	for p := range g.Graphs() {
		pairs = append(pairs, p)
	}
	return pairs
}

func (g *Gofer) Ticks(pairs ...graph.Pair) ([]graph.IndirectTick, error) {
	var ticks []graph.IndirectTick
	for _, pair := range pairs {
		if pairGraph, ok := g.graphs[pair]; ok {
			g.ingestor.Feed(pairGraph)
			ticks = append(ticks, pairGraph.Tick())
		} else {
			return nil, fmt.Errorf("unable to find %s pair", pair)
		}
	}

	return ticks, nil
}

func (g *Gofer) Origins(pairs ...graph.Pair) (map[graph.Pair][]string, error) {
	origins := map[graph.Pair][]string{}
	for _, pair := range pairs {
		if pairGraph, ok := g.graphs[pair]; ok {
			graph.Walk(pairGraph, func(node graph.Node) {
				if originNode, ok := node.(*graph.OriginNode); ok {
					name := originNode.OriginPair().Origin
					for _, n := range origins[pair] {
						if name == n {
							return
						}
					}
					origins[pair] = append(origins[pair], name)
				}
			})
		} else {
			return nil, fmt.Errorf("unable to find %s pair", pair)
		}
	}

	return origins, nil
}
