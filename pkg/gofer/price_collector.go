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
	"github.com/makerdao/gofer/pkg/exchange"
	"github.com/makerdao/gofer/pkg/model"
)

// PriceCollector will collect prices for you
type PriceCollector struct {
	exchangeSet *exchange.Set
}

// NewPriceCollector create new ready to work `PriceCollector`
func NewPriceCollector(set *exchange.Set) *PriceCollector {
	return &PriceCollector{
		exchangeSet: set,
	}
}

// CollectPricePoint makes request to exchange and fetching a price point
func (pc *PriceCollector) CollectPricePoint(pp *model.PotentialPricePoint) (*model.PricePoint, error) {
	return pc.exchangeSet.Call(pp)
}
