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
	aggregate *model.PriceAggregate
}

func NewTrade(pair *model.Pair) *Trade {
	return &Trade{
		aggregate: model.NewPriceAggregate("trade", &model.PricePoint{Pair: pair}),
	}
}

func (t *Trade) Ingest(pa *model.PriceAggregate) {
	if t.aggregate.Price == 0 {
		t.aggregate.Price = pa.Price
	} else {
		t.aggregate.Price *= pa.Price
	}
	t.aggregate.Prices = append(t.aggregate.Prices, pa)
}

func (t *Trade) Aggregate(pair *model.Pair) *model.PriceAggregate {
	if pair == nil || !pair.Equal(t.aggregate.Pair) {
		return nil
	}
  return t.aggregate.Clone()
}
