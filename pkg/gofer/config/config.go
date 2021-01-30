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

package config

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/makerdao/gofer/pkg/gofer"
	"github.com/makerdao/gofer/pkg/gofer/feeder"
	"github.com/makerdao/gofer/pkg/gofer/graph"
	"github.com/makerdao/gofer/pkg/gofer/origins"
	"github.com/makerdao/gofer/pkg/log"
)

const defaultMaxTTL = 60 * time.Second
const minTTLDifference = 30 * time.Second

type ErrCyclicReference struct {
	Pair graph.Pair
	Path []graph.Node
}

func (e ErrCyclicReference) Error() string {
	s := strings.Builder{}
	s.WriteString(fmt.Sprintf("a cyclic reference was detected for the %s pair: ", e.Path))
	for i, n := range e.Path {
		t := reflect.TypeOf(n).String()
		switch typedNode := n.(type) {
		case graph.Aggregator:
			s.WriteString(fmt.Sprintf("%s(%s)", t, typedNode.Pair()))
		default:
			s.WriteString(t)
		}
		if i != len(e.Path)-1 {
			s.WriteString(" -> ")
		}
	}
	return s.String()
}

type Config struct {
	Origins     map[string]Origin     `json:"origins"`
	PriceModels map[string]PriceModel `json:"priceModels"`
}

type Origin struct {
	Type   string          `json:"type"`
	Name   string          `json:"name"`
	Params json.RawMessage `json:"params"`
}

type PriceModel struct {
	Method  string          `json:"method"`
	Sources [][]Source      `json:"sources"`
	Params  json.RawMessage `json:"params"`
	TTL     int             `json:"ttl"`
}

type MedianPriceModel struct {
	MinSourceSuccess int `json:"minimumSuccessfulSources"`
}

type Source struct {
	Origin string `json:"origin"`
	Pair   string `json:"pair"`
	TTL    int    `json:"ttl"`
}

type Dependencies struct {
	Logger log.Logger
}

type Instances struct {
	Feeder *feeder.Feeder
	Gofer  *gofer.Gofer
}

func (c *Config) Configure(deps Dependencies) (*Instances, error) {
	// Graphs:
	gra, err := c.buildGraphs()
	if err != nil {
		return nil, err
	}

	// Feeder:
	fed := feeder.NewFeeder(origins.DefaultSet(), deps.Logger)

	// Gofer:
	gof := gofer.NewGofer(gra, fed)

	return &Instances{
		Feeder: fed,
		Gofer:  gof,
	}, nil
}

func (c *Config) buildGraphs() (map[graph.Pair]graph.Aggregator, error) {
	var err error

	graphs := map[graph.Pair]graph.Aggregator{}

	// It's important to create root nodes before branches, because branches
	// may refer to another root nodes instances.
	err = c.buildRoots(graphs)
	if err != nil {
		return nil, err
	}

	err = c.buildBranches(graphs)
	if err != nil {
		return nil, err
	}

	err = c.detectCycle(graphs)
	if err != nil {
		return nil, err
	}

	return graphs, nil
}

func (c *Config) buildRoots(graphs map[graph.Pair]graph.Aggregator) error {
	for name, model := range c.PriceModels {
		modelPair, err := graph.NewPair(name)
		if err != nil {
			return err
		}

		switch model.Method {
		case "median":
			var params MedianPriceModel
			if model.Params != nil {
				err := json.Unmarshal(model.Params, &params)
				if err != nil {
					return err
				}
			}
			graphs[modelPair] = graph.NewMedianAggregatorNode(modelPair, params.MinSourceSuccess)
		default:
			return fmt.Errorf("unknown method: %s", model.Method)
		}
	}

	return nil
}

func (c *Config) buildBranches(graphs map[graph.Pair]graph.Aggregator) error {
	for name, model := range c.PriceModels {
		// We can ignore error here, because it was checked already
		// in buildRoots method.
		modelPair, _ := graph.NewPair(name)

		var parent graph.Parent
		if typedNode, ok := graphs[modelPair].(graph.Parent); ok {
			parent = typedNode
		} else {
			return fmt.Errorf(
				"%s must implement the graph.Parent interface",
				reflect.TypeOf(graphs[modelPair]).Elem().String(),
			)
		}

		for _, sources := range model.Sources {
			var children []graph.Node
			for _, source := range sources {
				var err error
				var node graph.Node

				if source.Origin == "." {
					node, err = c.reference(graphs, source)
					if err != nil {
						return err
					}
				} else {
					node, err = c.originNode(model, source)
					if err != nil {
						return err
					}
				}

				children = append(children, node)
			}

			// If there are provided multiple sources it means, that price
			// have to be calculated by using the graph.IndirectAggregatorNode.
			// Otherwise we can pass that graph.OriginNode directly to
			// the parent node.
			var node graph.Node
			if len(children) == 1 {
				node = children[0]
			} else {
				indirectAggregator := graph.NewIndirectAggregatorNode(modelPair)
				for _, c := range children {
					indirectAggregator.AddChild(c)
				}
				node = indirectAggregator
			}

			parent.AddChild(node)
		}
	}

	return nil
}

func (c *Config) reference(graphs map[graph.Pair]graph.Aggregator, source Source) (graph.Node, error) {
	sourcePair, err := graph.NewPair(source.Pair)
	if err != nil {
		return nil, err
	}

	if _, ok := graphs[sourcePair]; !ok {
		return nil, fmt.Errorf(
			"unable to find price model for the %s pair",
			sourcePair,
		)
	}

	return graphs[sourcePair].(graph.Node), nil
}

func (c *Config) originNode(model PriceModel, source Source) (graph.Node, error) {
	sourcePair, err := graph.NewPair(source.Pair)
	if err != nil {
		return nil, err
	}

	originPair := graph.OriginPair{
		Origin: source.Origin,
		Pair:   sourcePair,
	}

	ttl := defaultMaxTTL
	if model.TTL > 0 {
		ttl = time.Second * time.Duration(model.TTL)
	}
	if source.TTL > 0 {
		ttl = time.Second * time.Duration(source.TTL)
	}

	return graph.NewOriginNode(originPair, ttl-minTTLDifference, ttl), nil
}

func (c *Config) detectCycle(graphs map[graph.Pair]graph.Aggregator) error {
	for _, pair := range sortGraphs(graphs) {
		if path := graph.DetectCycle(graphs[pair]); len(path) > 0 {
			return ErrCyclicReference{Pair: pair, Path: path}
		}
	}

	return nil
}

func sortGraphs(graphs map[graph.Pair]graph.Aggregator) []graph.Pair {
	var pairs []graph.Pair
	for p := range graphs {
		pairs = append(pairs, p)
	}
	sort.SliceStable(pairs, func(i, j int) bool {
		return pairs[i].String() < pairs[j].String()
	})
	return pairs
}
