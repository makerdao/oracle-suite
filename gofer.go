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

func (g *Gofer) paths(pairs []*model.Pair) []*model.PricePath {
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
func (g *Gofer) Prices(pairs ...*model.Pair) (map[model.Pair]*model.PriceAggregate, error) {
	ppaths := g.paths(pairs)
	aggregator := g.config.NewAggregator(ppaths)
	_, ppps := pather.FilterPotentialPricePoints(ppaths, g.config.Sources)

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
	ppaths := g.paths(pairs)
	return *model.NewPricePathMap(ppaths)
}

// Exchanges returns a list of Exchanges that will be queried for the given
// pairs
func (g *Gofer) Exchanges(pairs ...*model.Pair) []*model.Exchange {
	exchanges := make(map[string]*model.Exchange)

	if pairs == nil {
		for _, ppp := range g.config.Sources {
			exchanges[ppp.Exchange.Name] = ppp.Exchange
		}
	} else {
		exchangeIndex := make(map[model.Pair][]*model.Exchange)
		for _, ppp := range g.config.Sources {
			pair := *ppp.Pair
			exchangeIndex[pair] = append(exchangeIndex[pair], ppp.Exchange)
		}

		ppaths := g.paths(pairs)
		for _, ppath := range ppaths {
			for _, pair := range *ppath {
				if es, ok := exchangeIndex[*pair]; ok {
					for _, e := range es {
						exchanges[e.Name] = e
					}
				}
			}
		}
	}

	var result []*model.Exchange
	for _, e := range exchanges {
		result = append(result, e)
	}

	return result
}

// Pairs returns a list of pairs that are possible to resolve given the current
// configured Pather and set of PotentialPricePoints
func (g *Gofer) Pairs() []*model.Pair {
	pairs := g.config.Pather.Pairs()
	paths := g.paths(pairs)
	ppaths, _ := pather.FilterPotentialPricePoints(paths, g.config.Sources)
	targets := make(map[model.Pair]bool)
	for _, ppath := range ppaths {
		targets[*ppath.Target()] = true
	}
	var results []*model.Pair
	for pair := range targets {
		results = append(results, &pair)
	}
	return results
}
