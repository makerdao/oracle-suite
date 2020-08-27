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
	"github.com/makerdao/gofer/pkg/model"
)

func newTestPricePointAggregate(timestamp int64, exchange string, base string, quote string, price float64, volume float64) *model.PriceAggregate {
	return &model.PriceAggregate{
		PricePoint: &model.PricePoint{
			Timestamp: timestamp,
			Exchange:  &model.Exchange{Name: exchange},
			Pair:      &model.Pair{Base: base, Quote: quote},
			Price:     price,
			Ask:       price,
			Bid:       price,
			Volume:    volume,
		},
	}
}
