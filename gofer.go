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
	"makerdao/gofer/model"
	"makerdao/gofer/pather"
)

// Gofer library API
type Gofer struct {
	config    *Config
}

// NewGofer creates a new instance of the Gofer library API given a config
func NewGofer(config *Config) *Gofer {
	return &Gofer{
		config:    config,
	}
}

func (g *Gofer) paths(pairs ...*model.Pair) []*model.PricePath {
	var ppaths []*model.PricePath
	for _, pair := range pairs {
		ppaths_ := g.config.Pather.Path(pair)
		if ppaths_ != nil {
			ppaths = append(ppaths, ppaths_...)
		}
	}
	return ppaths
}

// Price returns a map of aggregated prices according
func (g *Gofer) Price(pairs ...*model.Pair) (map[model.Pair]*model.PriceAggregate, error) {
	ppaths := g.paths(pairs...)
	aggregator := g.config.NewAggregator(ppaths)
	ppps := pather.FilterPotentialPricePoints(ppaths, g.config.Sources)

	if _, err := g.config.Processor.Process(ppps, aggregator); err != nil {
		return nil, err
	}

	prices := make(map[model.Pair]*model.PriceAggregate)
	for _, pair := range pairs {
		prices[*pair] = aggregator.Aggregate(pair)
	}

	return prices, nil
}

// Paths returns a map of price paths for the given indirect pairs
func (g *Gofer) Paths(pairs ...*model.Pair) map[model.Pair][]*model.PricePath {
	ppaths := g.paths(pairs...)
	return *model.NewPricePathMap(ppaths)
}

// TODO: Implement getting configured exchanges
//func (g *Gofer) Exchanges(pair *model.Pair) []*model.Exchange {
//	return nil
//}

// TODO: Implement getting configured indirect pairs
//func (g *Gofer) Pairs() []*model.Pair {
//	return nil
//}
