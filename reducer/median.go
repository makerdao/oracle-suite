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
	"sort"

	"makerdao/gofer/model"
)

type MedianReducer struct {
	timeWindow      int64
	aggregate       *model.PriceAggregate
	prices          map[string]*model.PricePoint
	newestTimestamp int64
	reduced         bool
}

// Get price estimate from price point
func calcPrice(pp *model.PricePoint) uint64 {
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

func NewMedianReducer(pair *model.Pair, timeWindow int64) *MedianReducer {
	return &MedianReducer{
		timeWindow:      timeWindow,
		newestTimestamp: 0,
		prices:          make(map[string]*model.PricePoint),
		aggregate:       model.NewPriceAggregate(pair),
		reduced:         false,
	}
}

// Add a price point to median reducer state
func (r *MedianReducer) Ingest(pp *model.PricePoint) {
	// Ignore price point if asset pair not matching price model pair
	if !pp.Pair.Equal(r.aggregate.Pair) {
		return
	}

	// Ignore price point if no valid price
	price := calcPrice(pp)
	if price == 0 {
		return
	}

	if len(r.prices) == 0 || pp.Timestamp > r.newestTimestamp {
		r.newestTimestamp = pp.Timestamp
	}

	timeWindow := r.newestTimestamp - r.timeWindow
	// New price is outside time window, do nothing
	if pp.Timestamp <= timeWindow {
		return
	}

	existingPrice := r.prices[pp.Exchange.Name]
	// Price with same exchange as new price already exists
	if existingPrice == nil || pp.Timestamp > existingPrice.Timestamp {
		// Update existing price if new price is newer
		r.prices[pp.Exchange.Name] = pp
		// Set state to dirty
		r.reduced = false
	}
}

// Sort prices in state and return median
func (r *MedianReducer) Reduce() *model.PriceAggregate {
	if r.reduced || len(r.prices) == 0 {
		return r.aggregate.Clone()
	}

	timeWindow := r.newestTimestamp - r.timeWindow
	var prices []*model.PricePoint
	for _, p := range r.prices {
		// Only add prices inside time window
		if p.Timestamp > timeWindow {
			prices = append(prices, p)
		} else {
			delete(r.prices, p.Exchange.Name)
		}
	}
	priceCount := len(prices)

	// Sort price points by price
	sort.Slice(prices, func(i, j int) bool {
		return calcPrice(prices[i]) > calcPrice(prices[j])
	})

	if priceCount%2 == 0 {
		// Even price point count, take the mean of the two middle prices
		i := int(priceCount / 2)
		price1 := calcPrice(prices[i-1])
		price2 := calcPrice(prices[i])
		r.aggregate.Price = uint64((price1 + price2) / 2)
	} else {
		// Odd price point count, use the middle price
		i := int((priceCount - 1) / 2)
		r.aggregate.Price = calcPrice(prices[i])
	}
	r.aggregate.Prices = prices
	r.reduced = true
	return r.aggregate.Clone()
}
