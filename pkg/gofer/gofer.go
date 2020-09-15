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
	ingestor *graph.Ingestor
}

func NewGofer(graphs map[graph.Pair]graph.Aggregator, ingestor *graph.Ingestor) *Gofer {
	return &Gofer{
		graphs:   graphs,
		ingestor: ingestor,
	}
}

func (g *Gofer) Graphs() map[graph.Pair]graph.Aggregator {
	return g.graphs
}

func (g *Gofer) Ingestor() *graph.Ingestor {
	return g.ingestor
}

func (g *Gofer) Ticks(pairs ...graph.Pair) ([]graph.IndirectTick, error) {
	var ticks []graph.IndirectTick
	for _, pair := range pairs {
		if pairGraph, ok := g.graphs[pair]; ok {
			g.ingestor.Ingest(pairGraph)
			ticks = append(ticks, pairGraph.Tick())
		} else {
			return nil, fmt.Errorf("unable to find %s pair", pair)
		}
	}

	return ticks, nil
}

func (g *Gofer) Exchanges(pairs ...graph.Pair) (map[graph.Pair][]string, error) {
	exchanges := map[graph.Pair][]string{}
	for _, pair := range pairs {
		if pairGraph, ok := g.graphs[pair]; ok {
			graph.Walk(pairGraph, func(node graph.Node) {
				if exchangeNode, ok := node.(*graph.ExchangeNode); ok {
					name := exchangeNode.ExchangePair().Exchange
					for _, n := range exchanges[pair] {
						if name == n {
							return
						}
					}
					exchanges[pair] = append(exchanges[pair], name)
				}
			})
		} else {
			return nil, fmt.Errorf("unable to find %s pair", pair)
		}
	}

	return exchanges, nil
}
