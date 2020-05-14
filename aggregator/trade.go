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
)

type Trade struct {
	ppath model.PricePath
	aggregate *model.PriceAggregate
}

func NewTrade() *Trade {
	return &Trade{
		aggregate: model.NewPriceAggregate("trade", &model.PricePoint{}),
	}
}

func (t *Trade) Ingest(next *model.PriceAggregate) {
	current := t.aggregate
	t.ppath = append(t.ppath, next.Pair)
	if current.Price == 0 {
		current.Price = next.Price
	} else if current.Pair.Base == next.Pair.Base {
		current.Price = next.Price / current.Price
	} else {
		current.Price *= next.Price
	}
	current.Pair = t.ppath.Target()
	current.Prices = append(current.Prices, next)
}

func (t *Trade) Aggregate(pair *model.Pair) *model.PriceAggregate {
	if pair == nil || len(t.ppath) == 0 || !pair.Equal(t.ppath.Target()) {
		return nil
	}
	return t.aggregate.Clone()
}
