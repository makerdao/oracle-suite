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

	"github.com/makerdao/gofer/pkg/gofer/feeder"
	"github.com/makerdao/gofer/pkg/gofer/graph"
)

type ErrPairNotFound struct {
	AssetPair string
}

func (e ErrPairNotFound) Error() string {
	return fmt.Sprintf("unable to find the %s pair", e.AssetPair)
}

type Gofer struct {
	graphs map[graph.Pair]graph.Aggregator
	feeder *feeder.Feeder
}

// NewGofer returns a new Gofer instance.
func NewGofer(graphs map[graph.Pair]graph.Aggregator, feeder *feeder.Feeder) *Gofer {
	return &Gofer{
		graphs: graphs,
		feeder: feeder,
	}
}

// Graphs returns a node's graph used by this Gofer instance.
func (g *Gofer) Graphs() map[graph.Pair]graph.Aggregator {
	return g.graphs
}

// Feeder returns a feeder used by this Gofer instance.
func (g *Gofer) Feeder() *feeder.Feeder {
	return g.feeder
}

// Pairs returns the list of all supported asset pairs.
func (g *Gofer) Pairs() []graph.Pair {
	var ps []graph.Pair
	for p := range g.graphs {
		ps = append(ps, p)
	}
	return ps
}

// Feed feeds given asset pairs with prices using Feeder.
func (g *Gofer) Feed(pairs ...graph.Pair) (feeder.Warnings, error) {
	nodes, err := g.getNodesForPairs(pairs)
	if err != nil {
		return feeder.Warnings{}, err
	}
	return g.feeder.Feed(nodes), nil
}

// StartFeeder starts Feeder process that will automatically be updating prices
// in the background.
func (g *Gofer) StartFeeder(pairs ...graph.Pair) error {
	nodes, err := g.getNodesForPairs(pairs)
	if err != nil {
		return err
	}
	return g.feeder.Start(nodes)
}

// StopFeeder stops Feeder previously started with the StartFeeder method.
func (g *Gofer) StopFeeder() {
	g.feeder.Stop()
}

// Tick returns a Tick for the given pair.
func (g *Gofer) Tick(pair graph.Pair) (graph.AggregatorTick, error) {
	if node, ok := g.graphs[pair]; ok {
		return node.Tick(), nil
	}
	return graph.AggregatorTick{}, ErrPairNotFound{AssetPair: pair.String()}
}

// Ticks returns a list of Ticks for the given pairs.
func (g *Gofer) Ticks(pairs ...graph.Pair) ([]graph.AggregatorTick, error) {
	var ticks []graph.AggregatorTick
	for _, pair := range pairs {
		if pg, ok := g.graphs[pair]; ok {
			ticks = append(ticks, pg.Tick())
		} else {
			return nil, ErrPairNotFound{AssetPair: pair.String()}
		}
	}
	return ticks, nil
}

// Origins returns all origins names which are involved in the calculation
// of the price for the given pairs.
func (g *Gofer) Origins(pairs ...graph.Pair) (map[graph.Pair][]string, error) {
	origins := map[graph.Pair][]string{}
	for _, pair := range pairs {
		if pg, ok := g.graphs[pair]; ok {
			graph.Walk(func(node graph.Node) {
				if on, ok := node.(*graph.OriginNode); ok {
					name := on.OriginPair().Origin
					for _, n := range origins[pair] {
						if name == n {
							return
						}
					}
					origins[pair] = append(origins[pair], name)
				}
			}, pg)
		} else {
			return nil, ErrPairNotFound{AssetPair: pair.String()}
		}
	}
	return origins, nil
}

func (g *Gofer) getNodesForPairs(pairs []graph.Pair) ([]graph.Node, error) {
	var graphs []graph.Node
	for _, pair := range pairs {
		if pg, ok := g.graphs[pair]; ok {
			graphs = append(graphs, pg)
		} else {
			return nil, ErrPairNotFound{AssetPair: pair.String()}
		}
	}
	return graphs, nil
}
