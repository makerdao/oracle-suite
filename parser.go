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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/makerdao/gofer/aggregator"
	"github.com/makerdao/gofer/exchange"
	"github.com/makerdao/gofer/model"
)

type AggregateProcessor interface {
	Process(pairs []*model.Pair, agg aggregator.Aggregator) (aggregator.Aggregator, error)
}

type Config struct {
	//Exchanges  []model.Exchange            `json:"exchanges"`
	Aggregator aggregator.AggregatorParams `json:"aggregator"`
	// TODO: add Processor params
}

func FromJSON(b []byte) (*Gofer, error) {
	var jc Config
	if err := json.Unmarshal(b, &jc); err != nil {
		return nil, fmt.Errorf("failed to parse config from JSON: %w", err)
	}

	agg, err := aggregator.FromConfig(jc.Aggregator)
	if err != nil {
		return nil, fmt.Errorf("failed to create aggregator when parsing config: %w", err)
	}

	return NewGofer(agg, NewProcessor(exchange.DefaultExchangesSet())), nil
}

func ReadFile(path string) (*Gofer, error) {
	jsonFile, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load json config file: %w", err)
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	return FromJSON(byteValue)
}
