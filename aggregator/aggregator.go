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

import "github.com/makerdao/gofer/model"

type Aggregator interface {
	// Add a price point to be aggregated
	Ingest(*model.PriceAggregate)
	// Calculate asset pair aggregate returning nil if pair not available
	Aggregate(*model.Pair) *model.PriceAggregate
}

// Get price estimate from price point
func calcPrice(pp *model.PriceAggregate) float64 {
	// If ask/bid values are valid return mean of ask and bid
	if pp.Ask != 0 && pp.Bid != 0 {
		return (pp.Ask + pp.Bid) / 2
	}
	// If last auction price is valid return that
	if pp.Price != 0 {
		return pp.Price
	}
	// Otherwise return invalid price
	return 0
}
