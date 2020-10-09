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
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/makerdao/gofer/pkg/graph"
)

const defaultMaxTTL = 60 * time.Second
const minTTLDifference = 30 * time.Second

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
	Sources [][]JSONSource  `json:"sources"`
	Params  json.RawMessage `json:"params"`
	TTL     int             `json:"ttl"`
}

type MedianPriceModel struct {
	MinSourceSuccess int `json:"minimumSuccessfulSources"`
}

type JSONSource struct {
	Origin string `json:"origin"`
	Pair   string `json:"pair"`
	TTL    int    `json:"ttl"`
}

type JSONConfigErr struct {
	Err error
}

func (e JSONConfigErr) Error() string {
	return e.Err.Error()
}

func ParseJSONFile(path string) (*JSON, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load JSON config file: %w", err)
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, JSONConfigErr{fmt.Errorf("failed to load JSON config file: %w", err)}
	}

	return ParseJSON(b)
}

func ParseJSON(b []byte) (*JSON, error) {
	j := &JSON{}
	err := json.Unmarshal(b, j)
	if err != nil {
		return nil, JSONConfigErr{err}
	}

	return j, nil
}

func (j *JSON) BuildGraphs() (map[graph.Pair]graph.Aggregator, error) {
	var err error

	graphs := map[graph.Pair]graph.Aggregator{}

	// It's important to create root nodes before branches, because branches
	// may refer to another root nodes instances.
	err = j.buildRoots(graphs)
	if err != nil {
		return nil, JSONConfigErr{err}
	}

	err = j.buildBranches(graphs)
	if err != nil {
		return nil, JSONConfigErr{err}
	}

	err = j.detectCycle(graphs)
	if err != nil {
		return nil, JSONConfigErr{err}
	}

	return graphs, nil
}

func (j *JSON) buildRoots(graphs map[graph.Pair]graph.Aggregator) error {
	for name, model := range j.PriceModels {
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

func (j *JSON) buildBranches(graphs map[graph.Pair]graph.Aggregator) error {
	for name, model := range j.PriceModels {
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
					node, err = j.reference(graphs, source)
					if err != nil {
						return err
					}
				} else {
					node, err = j.originNode(model, source)
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

func (j *JSON) reference(graphs map[graph.Pair]graph.Aggregator, source JSONSource) (graph.Node, error) {
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

func (j *JSON) originNode(model JSONPriceModel, source JSONSource) (graph.Node, error) {
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

func (j *JSON) detectCycle(graphs map[graph.Pair]graph.Aggregator) error {
	for _, p := range sortGraphs(graphs) {
		if c := graph.DetectCycle(graphs[p]); len(c) > 0 {
			errMsg := strings.Builder{}
			errMsg.WriteString(fmt.Sprintf("cyclic reference was detected for the %s pair: ", p))
			for i, n := range c {
				switch typedNode := n.(type) {
				case *graph.MedianAggregatorNode:
					errMsg.WriteString("median:" + typedNode.Pair().String())
				case *graph.IndirectAggregatorNode:
					errMsg.WriteString("indirect:" + typedNode.Pair().String())
				default:
					errMsg.WriteString(reflect.TypeOf(typedNode).String())
				}
				if i != len(c)-1 {
					errMsg.WriteString(" -> ")
				}
			}
			return errors.New(errMsg.String())
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
