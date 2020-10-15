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

type PriceModels map[graph.Pair]graph.Aggregator

func (g PriceModels) Pairs() []graph.Pair {
	var pairs []graph.Pair
	for p := range g {
		pairs = append(pairs, p)
	}
	return pairs
}

func (g PriceModels) Origins(pairs ...graph.Pair) (map[graph.Pair][]string, error) {
	origins := map[graph.Pair][]string{}
	for _, pair := range pairs {
		if pairGraph, ok := g[pair]; ok {
			graph.Walk(func(node graph.Node) {
				if originNode, ok := node.(*graph.OriginNode); ok {
					name := originNode.OriginPair().Origin
					for _, n := range origins[pair] {
						if name == n {
							return
						}
					}
					origins[pair] = append(origins[pair], name)
				}
			}, pairGraph)
		} else {
			return nil, fmt.Errorf("unable to find %s pair", pair)
		}
	}
	return origins, nil
}

func (g PriceModels) Ticks(pairs ...graph.Pair) ([]graph.AggregatorTick, error) {
	var ticks []graph.AggregatorTick
	for _, pair := range pairs {
		ticks = append(ticks, g[pair].Tick())
	}
	return ticks, nil
}

func AllNodes(g PriceModels) []graph.Node {
	var nodes []graph.Node
	for _, pairGraph := range g {
		nodes = append(nodes, pairGraph)
	}
	return nodes
}
