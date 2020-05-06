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
	"makerdao/gofer/model"
	"makerdao/gofer/pather"
)

type Path struct {
	pather             pather.Pather
	newDirectReducer   func(*model.Pair) Aggregator
	newIndirectReducer func(*model.Pair) Aggregator
	newTradeReducer    func(*model.Pair) Aggregator
	reducers           map[model.Pair]Aggregator
}

func NewPath(pather pather.Pather, newDirectReducer func(*model.Pair) Aggregator, newIndirectReducer func(*model.Pair) Aggregator, newTradeReducer func(*model.Pair) Aggregator) *Path {
	return &Path{
		pather:             pather,
		newDirectReducer:   newDirectReducer,
		newIndirectReducer: newIndirectReducer,
		newTradeReducer:    newTradeReducer,
		reducers:           make(map[model.Pair]Aggregator),
	}
}

func NewPathWithDefaultTrade(pather pather.Pather, newDirectReducer func(*model.Pair) Aggregator, newIndirectReducer func(*model.Pair) Aggregator) *Path {
	return NewPath(
		pather,
		newDirectReducer,
		newIndirectReducer,
		func(pair *model.Pair) Aggregator {
			return NewTrade(pair)
		},
	)
}

func (r *Path) resolve(ppath model.PricePath) *model.PriceAggregate {
	target := ppath.Target()
	trade := r.newTradeReducer(target)
	//pa := model.NewPriceAggregate(ppath.String(), &model.PricePoint{Pair: ppath.Target()})
	for _, pair := range ppath {
		if r_, ok := r.reducers[*pair]; ok {
			trade.Ingest(r_.Aggregate(pair))
		} else {
			return nil
		}
	}
	return trade.Aggregate(target)
}

func (r *Path) Ingest(pa *model.PriceAggregate) {
	pair := *pa.Pair
	if _, ok := r.reducers[pair]; !ok {
		r.reducers[pair] = r.newDirectReducer(pa.Pair)
	}
	r.reducers[pair].Ingest(pa)
}

func (r *Path) Aggregate(pair *model.Pair) *model.PriceAggregate {
	ppaths := r.pather.Path(pair)
	if ppaths == nil {
		return nil
	}

	topReducer := r.newIndirectReducer(pair)
	for _, path := range ppaths.Paths {
		if pa := r.resolve(path); pa != nil {
			topReducer.Ingest(pa)
		}
	}

	return topReducer.Aggregate(pair)
}
