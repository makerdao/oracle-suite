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

	. "github.com/makerdao/gofer"
	. "github.com/makerdao/gofer/aggregator"
	. "github.com/makerdao/gofer/model"
	. "github.com/makerdao/gofer/pather"
)

type jsonSource struct {
	Base       string            `json:"base"`
	Quote      string            `json:"quote"`
	Exchange   string            `json:"exchange"`
	Parameters map[string]string `json:"parameters"`
}

type jsonAggregator struct {
	Name       string                 `json:"name"`
	Parameters map[string]interface{} `json:"parameters"`
}

type jsonPather struct {
	Name       string                 `json:"name"`
	Parameters map[string]interface{} `json:"parameters"`
}

type jsonConfig struct {
	Sources    []jsonSource   `json:"sources"`
	Aggregator jsonAggregator `json:"aggregator"`
	Pather     jsonPather     `json:"pather"`
}

func patherFromJSON(jp *jsonPather) (Pather, error) {
	switch jp.Name {
	case "setzer":
		return NewSetzer(), nil
	}
	return nil, fmt.Errorf("no pather found with name %s", jp.Name)
}

func aggregatorFromJSON(jc *jsonConfig, ja *jsonAggregator) (func([]*PricePath) Aggregator, error) {
	switch ja.Name {
	case "median":
		// JSON numbers will always be parsed to float64
		timewindow, ok := ja.Parameters["timewindow"].(float64)
		if !ok {
			return nil, fmt.Errorf("couldn't parse median aggregator parameter: timewindow as number")
		}

		return func(ppaths []*PricePath) Aggregator {
			return NewMedian(int64(timewindow))
		}, nil

	case "path":
		jaDirect, ok := ja.Parameters["direct"].(jsonAggregator)
		if !ok {
			return nil, fmt.Errorf("couldn't parse path aggregator parameter: direct as aggregator")
		}
		newDirect, err := aggregatorFromJSON(jc, &jaDirect)
		if err != nil {
			return nil, err
		}

		return func(ppaths []*PricePath) Aggregator {
			return NewPath(ppaths, newDirect(ppaths))
		}, nil
	}

	return nil, fmt.Errorf("no aggregator found with name \"%s\"", ja.Name)
}

func fromJSON(jc *jsonConfig) (*Config, error) {
	var ppps []*PotentialPricePoint
	for _, s := range jc.Sources {
		ppp := &PotentialPricePoint{
			Pair: &Pair{
				Base:  s.Base,
				Quote: s.Quote,
			},
			Exchange: &Exchange{
				Name:   s.Exchange,
				Config: s.Parameters,
			},
		}
		ppps = append(ppps, ppp)
	}

	newAggregator, err := aggregatorFromJSON(jc, &jc.Aggregator)
	if err != nil {
		return nil, err
	}

	pather, err := patherFromJSON(&jc.Pather)
	if err != nil {
		return nil, err
	}

	return NewConfig(ppps, newAggregator, pather, nil), nil
}

func ReadConfig(blob []byte) (*Config, error) {
	var jc jsonConfig
	if err := json.Unmarshal(blob, &jc); err != nil {
		return nil, err
	}

	config, err := fromJSON(&jc)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func ReadFile(path string) (*Config, error) {
	jsonFile, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load json config file: %w", err)
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	return ReadConfig(byteValue)
}
