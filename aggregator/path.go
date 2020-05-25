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

package aggregator

import (
	. "github.com/makerdao/gofer/model"
)

// Path is an aggregator that resolves price paths for indirect pairs and takes
// the median of all paths for each pair
type Path struct {
	paths            PricePathMap
	directAggregator Aggregator
}

// NewPath returns a new instance of `Path` that uses the given price paths to
// aggregate indirect pairs and an aggregator to merge direct pairs.
func NewPath(ppaths []*PricePath, directAggregator Aggregator) *Path {
	return &Path{
		paths:            *NewPricePathMap(ppaths),
		directAggregator: directAggregator,
	}
}

// Calculate the final trade price of an ordered list of prices
func trade(pas []*PriceAggregate) *PriceAggregate {
	var pair *Pair
	var price float64

	for _, pa := range pas {
		if price == 0 {
			price = pa.Price
			pair = pa.Pair.Clone()
		} else if pair.Base == pa.Pair.Base {
			price = pa.Price / price
			pair.Base = pair.Quote
			pair.Quote = pa.Pair.Quote
		} else {
			price *= pa.Price
			pair.Quote = pa.Pair.Quote
		}
	}

	return NewPriceAggregate(
		"trade",
		&PricePoint{
			Pair:  pair,
			Price: price,
		},
		pas...,
	)
}

func (r *Path) resolve(ppath PricePath) *PriceAggregate {
	var pas []*PriceAggregate
	for _, pair := range ppath {
		pa := r.directAggregator.Aggregate(pair)
		if pa == nil {
			return nil
		}

		pas = append(pas, pa)
	}
	return trade(pas)
}

func (r *Path) Ingest(pa *PriceAggregate) {
	r.directAggregator.Ingest(pa)
}

func (r *Path) Aggregate(pair *Pair) *PriceAggregate {
	if pair == nil {
		return nil
	}

	ppaths := r.paths[*pair]
	if ppaths == nil {
		return nil
	}

	var pas []*PriceAggregate
	var prices []float64
	for _, path := range ppaths {
		if pa := r.resolve(*path); pa != nil {
			pas = append(pas, pa)
			prices = append(prices, pa.Price)
		}
	}

	return NewPriceAggregate(
		"path",
		&PricePoint{
			Pair:  pair.Clone(),
			Price: median(prices),
		},
		pas...,
	)
}
