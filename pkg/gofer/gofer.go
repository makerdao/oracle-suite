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
	"fmt"
	"log"

	"github.com/makerdao/gofer/pkg/model"
)

type Aggregator interface {
	// Return aggregated asset pair returning nil if pair not available
	Aggregate(*model.Pair) *model.PriceAggregate
}

type Ingestor interface {
	// Informs clients of the required data points
	GetSources(...*model.Pair) []*model.PricePoint
	// Allows for feeding data points to the IngestingAggregator
	Ingest(*model.PriceAggregate)
}

type Fetcher interface {
	Fetch([]*model.PricePoint)
}

// An IngestingAggregator is a service that, when fed the required data points,
// is able to return aggregated information derived from them in a way specific
// to the embedded aggregation model.
type IngestingAggregator interface {
	Aggregator
	Ingestor
}

func NewGofer(a IngestingAggregator, f Fetcher) *Gofer {
	return &Gofer{
		aggregator: a,
		fetcher:    f,
	}
}

type Gofer struct {
	aggregator IngestingAggregator
	fetcher    Fetcher
}

func (g *Gofer) Prices(pairs ...*model.Pair) (map[model.Pair]*model.PriceAggregate, error) {
	f := func(agg IngestingAggregator, pairs ...*model.Pair) error {
		if agg == nil {
			return fmt.Errorf("no working agregator passed to processor")
		}

		sources := agg.GetSources(pairs...)
		g.fetcher.Fetch(sources)

		for _, pp := range sources {
			if pp.Error != nil {
				log.Println(pp.Error)
				continue
			}

			pa := &model.PriceAggregate{
				PriceModelName: fmt.Sprintf("exchange[%s]", pp.Exchange.Name),
				PricePoint:     pp,
			}

			agg.Ingest(pa)
		}

		return nil
	}

	if err := f(g.aggregator, pairs...); err != nil {
		return nil, err
	}

	prices := make(map[model.Pair]*model.PriceAggregate)
	for _, pair := range pairs {
		prices[*pair] = g.aggregator.Aggregate(pair)
	}

	return prices, nil
}

func (g *Gofer) Exchanges(pairs ...*model.Pair) []*model.Exchange {
	exchanges := make(map[string]*model.Exchange)

	for _, ppp := range g.aggregator.GetSources(pairs...) {
		exchanges[ppp.Exchange.Name] = ppp.Exchange
	}

	result := make([]*model.Exchange, 0)
	for _, e := range exchanges {
		result = append(result, e)
	}

	return result
}

func (g *Gofer) Pairs() []*model.Pair {
	pairs := make(map[string]*model.Pair)

	for _, ppp := range g.aggregator.GetSources() {
		pairs[ppp.Pair.String()] = ppp.Pair
	}

	result := make([]*model.Pair, 0)
	for _, p := range pairs {
		result = append(result, p)
	}

	return result
}
