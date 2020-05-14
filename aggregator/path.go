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
	. "makerdao/gofer/model"
)

type Path struct {
	paths                 map[Pair]*PricePaths
	newDirectAggregator   func(*Pair) Aggregator
	newIndirectAggregator func(*Pair) Aggregator
	aggregators           map[Pair]Aggregator
}

func NewPath(ppaths []*PricePaths, newDirectAggregator func(*Pair) Aggregator, newIndirectAggregator func(*Pair) Aggregator) *Path {
	paths := make(map[Pair]*PricePaths)
	for _, ppath := range ppaths {
		paths[*ppath.Target] = ppath
	}
	return &Path{
		paths:                 paths,
		newDirectAggregator:   newDirectAggregator,
		newIndirectAggregator: newIndirectAggregator,
		aggregators:           make(map[Pair]Aggregator),
	}
}

func trade(pas []*PriceAggregate) *PriceAggregate {
	var pair *Pair
	var price uint64

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
		if r_, ok := r.aggregators[*pair]; ok {
			pas = append(pas, r_.Aggregate(pair))
		} else {
			return nil
		}
	}
	return trade(pas)
}

func (r *Path) Ingest(pa *PriceAggregate) {
	pair := *pa.Pair
	if _, ok := r.aggregators[pair]; !ok {
		r.aggregators[pair] = r.newDirectAggregator(pa.Pair)
	}
	r.aggregators[pair].Ingest(pa)
}

func (r *Path) Aggregate(pair *Pair) *PriceAggregate {
	ppaths := r.paths[*pair]
	if ppaths == nil {
		return nil
	}

	rootAggregator := r.newIndirectAggregator(pair)
	for _, path := range ppaths.Paths {
		if pa := r.resolve(path); pa != nil {
			rootAggregator.Ingest(pa)
		}
	}

	return rootAggregator.Aggregate(pair)
}
