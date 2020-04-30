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

package reducer

import (
	"math/rand"

	"makerdao/gofer/model"
)

func RandomReduce(r Reducer, prices []*model.PricePoint) *model.PriceAggregate {
	for _, i := range rand.Perm(len(prices)) {
		r.Ingest(prices[i])
		if rand.Intn(2) == 1 {
			r.Reduce()
		}
	}
	return r.Reduce()
}

func NewTestPricePoint(timestamp int64, exchange string, base string, quote string, price uint64, volume uint64) *model.PricePoint {
	return &model.PricePoint{
		Timestamp: timestamp,
		Exchange:  &model.Exchange{Name: exchange},
		Pair:      &model.Pair{Base: base, Quote: quote},
		Price:     price,
		Ask:       price,
		Bid:       price,
		Volume:    volume,
	}
}

func NewTestPricePointPriceOnly(timestamp int64, exchange string, base string, quote string, last uint64, volume uint64) *model.PricePoint {
	return &model.PricePoint{
		Timestamp: timestamp,
		Exchange:  &model.Exchange{Name: exchange},
		Pair:      &model.Pair{Base: base, Quote: quote},
		Price:     last,
		Volume:    volume,
	}
}
