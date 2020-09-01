package config

import "encoding/json"

type JSON struct {
	Origins    Origin     `json:"origin"`
	Aggregator Aggregator `json:"aggregator"`
}

type Origin struct {
	Type   string      `json:"type"`
	Name   string      `json:"name"`
	Config interface{} `json:"config"`
}

type Sources struct {
	Origin string `json:"origin"`
	Pair   string `json:"pair"`
}

type PriceModel struct {
	Method           string      `json:"method"`
	MinSourceSuccess int         `json:"minSourceSuccess"`
	Sources          [][]Sources `json:"sources"`
}

type Parameters struct {
	PriceModels map[string]PriceModel `json:"pricemodels"`
}

type Aggregator struct {
	Name       string     `json:"name"`
	Parameters Parameters `json:"parameters"`
}

func ParseJSON(b []byte) (*JSON, error) {
	j := &JSON{}
	err := json.Unmarshal(b, j)
	if err != nil {
		return nil, err
	}

	return j, nil
}
