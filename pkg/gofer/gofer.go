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
	"github.com/makerdao/gofer/pkg/aggregator"
	"github.com/makerdao/gofer/pkg/model"
)

// Gofer library API
type Gofer struct {
	aggregator aggregator.Aggregator
	processor  AggregateProcessor
}

// NewGofer creates a new instance of the Gofer library API given a config
func NewGofer(agg aggregator.Aggregator, processor AggregateProcessor) *Gofer {
	return &Gofer{
		aggregator: agg,
		processor:  processor,
	}
}

// SetProcessor sets new `AggregateProcessor` to gofer instance
func (g *Gofer) SetProcessor(processor AggregateProcessor) {
	g.processor = processor
}

// Price returns a map of aggregated prices according
func (g *Gofer) Prices(pairs ...*model.Pair) (map[model.Pair]*model.PriceAggregate, error) {
	if _, err := g.processor.Process(pairs, g.aggregator); err != nil {
		return nil, err
	}

	prices := make(map[model.Pair]*model.PriceAggregate)
	for _, pair := range pairs {
		prices[*pair] = g.aggregator.Aggregate(pair)
	}

	return prices, nil
}

// Exchanges returns a list of Exchange that support all pairs
func (g *Gofer) Exchanges(pairs ...*model.Pair) []*model.Exchange {
	exchanges := make(map[string]*model.Exchange)

	// Get all exchanges for aggregator
	for _, ppp := range g.aggregator.GetSources(pairs) {
		exchanges[ppp.Exchange.Name] = ppp.Exchange
	}

	result := make([]*model.Exchange, 0)
	for _, exchange := range exchanges {
		result = append(result, exchange)
	}

	return result
}
