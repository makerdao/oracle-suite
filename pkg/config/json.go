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
	"io/ioutil"
	"os"
	"reflect"

	"github.com/makerdao/gofer/pkg/graph"
)

type JSON struct {
	Origins     map[string]JSONOrigin     `json:"origins"`
	PriceModels map[string]JSONPriceModel `json:"priceModels"`
}

type JSONOrigin struct {
	Type   string          `json:"type"`
	Name   string          `json:"name"`
	Params json.RawMessage `json:"params"`
}

type JSONPriceModel struct {
	Method  string          `json:"method"`
	Sources [][]JSONSources `json:"sources"`
	Params  json.RawMessage `json:"params"`
}

type MedianPriceModel struct {
	MinSourceSuccess int `json:"minimumSuccessfulSources"`
}

type JSONSources struct {
	Origin string `json:"origin"`
	Pair   string `json:"pair"`
}

func ParseJSONFile(path string) (*JSON, error) {
	f, err := os.Open(path)
	defer f.Close()

	if err != nil {
		return nil, fmt.Errorf("failed to load json config file: %w", err)
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to load json config file: %w", err)
	}

	return ParseJSON(b)
}

func ParseJSON(b []byte) (*JSON, error) {
	j := &JSON{}
	err := json.Unmarshal(b, j)
	if err != nil {
		return nil, err
	}

	return j, nil
}

func (j *JSON) BuildGraphs() (map[graph.Pair]graph.Aggregator, error) {
	var err error

	graphs := map[graph.Pair]graph.Aggregator{}

	err = j.buildRoots(graphs)
	if err != nil {
		return nil, err
	}

	err = j.buildBranches(graphs)
	if err != nil {
		return nil, err
	}

	return graphs, nil
}

func (j *JSON) buildRoots(graphs map[graph.Pair]graph.Aggregator) error {
	// It's important to create root nodes before branches,
	// because branches may refer to another root nodes
	// instances.
	for name, model := range j.PriceModels {
		pair, err := graph.NewPair(name)
		if err != nil {
			return err
		}

		switch model.Method {
		case "median":
			var params MedianPriceModel
			err := json.Unmarshal(model.Params, &params)
			if err != nil {
				return err
			}
			graphs[pair] = graph.NewMedianAggregatorNode(pair, params.MinSourceSuccess)
		default:
			return fmt.Errorf("unknown method: %s", model.Method)
		}
	}

	return nil
}

func (j *JSON) buildBranches(graphs map[graph.Pair]graph.Aggregator) error {
	origins := map[graph.OriginPair]graph.Origin{}

	for name, model := range j.PriceModels {
		pair, _ := graph.NewPair(name)

		var parent graph.Parent
		if typedNode, ok := graphs[pair].(graph.Parent); ok {
			parent = typedNode
		} else {
			return fmt.Errorf(
				"%s must implement graph.Parent interface",
				reflect.TypeOf(graphs[pair]).Elem().String(),
			)
		}

		for _, sources := range model.Sources {
			var children []graph.Node
			for _, source := range sources {
				sourcePair, err := graph.NewPair(source.Pair)
				if err != nil {
					return err
				}

				if source.Origin == "." {
					// The reference to an other root node.
					if _, ok := graphs[sourcePair]; !ok {
						return fmt.Errorf("unable to find price model for %s pair", sourcePair)
					}
					children = append(children, graphs[sourcePair].(graph.Node))
				} else {
					// The origin node. If it's possible we're trying to reuse
					// previously created origin nodes.
					pair, err := graph.NewPair(source.Pair)
					if err != nil {
						return err
					}
					originPair := graph.OriginPair{
						Origin: source.Origin,
						Pair:   pair,
					}
					if _, ok := origins[originPair]; !ok {
						origins[originPair] = graph.NewOriginNode(originPair)
					}
					children = append(children, origins[originPair])
				}
			}

			// If there are provided multiple sources it means, that price
			// have to be calculated by using graph.IndirectAggregatorNode.
			// Otherwise we can pass that graph.OriginNode directly to
			// the parent node.

			var node graph.Node
			if len(children) == 1 {
				node = children[0]
			} else {
				indirectAggregator := graph.NewIndirectAggregatorNode(pair)
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
