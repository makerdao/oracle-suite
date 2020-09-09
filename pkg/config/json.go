package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/makerdao/gofer/pkg/graph"
)

type JSON struct {
	Origins    JSONOrigin     `json:"origin"`
	Aggregator JSONAggregator `json:"aggregator"`
}

type JSONOrigin struct {
	Type   string      `json:"type"`
	Name   string      `json:"name"`
	Config interface{} `json:"config"`
}

type JSONAggregator struct {
	Name       string         `json:"name"`
	Parameters JSONParameters `json:"parameters"`
}

type JSONParameters struct {
	PriceModels map[string]JSONPriceModel `json:"pricemodels"`
}

type JSONPriceModel struct {
	Method           string          `json:"method"`
	MinSourceSuccess int             `json:"minSourceSuccess"`
	Sources          [][]JSONSources `json:"sources"`
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

func (j *JSON) BuildGraphs() (map[string]graph.Aggregator, error) {
	graphs := map[string]graph.Aggregator{}

	// Build roots:
	// It's important to do it before branches, because branches may refer to
	// another root nodes.
	for name, model := range j.Aggregator.Parameters.PriceModels {
		switch model.Method {
		case "median":
			graphs[name] = graph.NewMedianAggregatorNode(model.MinSourceSuccess)
		default:
			return nil, fmt.Errorf("unknown method: %s", model.Method)
		}
	}

	// Build branches:
	for name, model := range j.Aggregator.Parameters.PriceModels {
		for _, sources := range model.Sources {
			indirectAggregator := graph.NewIndirectAggregatorNode()

			for _, source := range sources {
				if source.Origin == "." {
					// The reference to an other root node:
					indirectAggregator.AddChild(
						graphs[source.Pair].(graph.Node),
					)
				} else {
					// The exchange node:
					pair, err := newPairFromString(source.Pair)
					if err != nil {
						return nil, err
					}

					exchangePair := graph.ExchangePair{
						Exchange: source.Origin,
						Pair:     pair,
					}

					indirectAggregator.AddChild(
						graph.NewExchangeNode(exchangePair),
					)
				}
			}

			switch typedNode := graphs[name].(type) {
			case *graph.MedianAggregatorNode:
				typedNode.AddChild(indirectAggregator)
			}
		}
	}

	return graphs, nil
}

func newPairFromString(s string) (graph.Pair, error) {
	ss := strings.Split(s, "/")
	if len(ss) != 2 {
		return graph.Pair{}, fmt.Errorf("couldn't parse pair \"%s\"", s)
	}

	if len(ss[0]) == 0 {
		return graph.Pair{}, fmt.Errorf("base asset name is empty")
	}

	if len(ss[1]) == 0 {
		return graph.Pair{}, fmt.Errorf("quote asset name is empty")
	}

	return graph.Pair{Base: strings.ToUpper(ss[0]), Quote: strings.ToUpper(ss[1])}, nil
}
