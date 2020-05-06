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
	"sort"

	"makerdao/gofer/model"
)

type Median struct {
	timeWindow      int64
	aggregate       *model.PriceAggregate
	prices          map[string]*model.PriceAggregate
	newestTimestamp int64
	reduced         bool
}

func NewMedian(pair *model.Pair, timeWindow int64) *Median {
	return &Median{
		timeWindow:      timeWindow,
		newestTimestamp: 0,
		prices:          make(map[string]*model.PriceAggregate),
		aggregate:       model.NewPriceAggregate("median", &model.PricePoint{Pair: pair}),
		reduced:         false,
	}
}

// Add a price point to median reducer state
func (r *Median) Ingest(pa *model.PriceAggregate) {
	// Ignore price point if asset pair not matching price model pair
	if !pa.Pair.Equal(r.aggregate.Pair) {
		return
	}

	// Ignore price point if no valid price
	price := calcPrice(pa)
	if price == 0 {
		return
	}

	if len(r.prices) == 0 || pa.Timestamp > r.newestTimestamp {
		r.newestTimestamp = pa.Timestamp
	}

	timeWindow := r.newestTimestamp - r.timeWindow
	// New price is outside time window, do nothing
	if pa.Timestamp <= timeWindow {
		return
	}

	existingPrice := r.prices[pa.Exchange.Name]
	// Price with same exchange as new price already exists
	if existingPrice == nil || pa.Timestamp > existingPrice.Timestamp {
		// Update existing price if new price is newer
		r.prices[pa.Exchange.Name] = pa
		// Set state to dirty
		r.reduced = false
	}
}

// Sort prices in state and return median
func (r *Median) Aggregate(pair *model.Pair) *model.PriceAggregate {
	if pair == nil || !pair.Equal(r.aggregate.Pair) {
		return nil
	}

	if r.reduced || len(r.prices) == 0 {
		return r.aggregate.Clone()
	}

	timeWindow := r.newestTimestamp - r.timeWindow
	var pas []*model.PriceAggregate
	for _, p := range r.prices {
		// Only add prices inside time window
		if p.Timestamp > timeWindow {
			pas = append(pas, p)
		} else {
			delete(r.prices, p.Exchange.Name)
		}
	}

	prices := make([]uint64, len(pas))
	for i, pa := range pas {
		prices[i] = calcPrice(pa)
	}
	r.aggregate.Price = median(prices)
	r.aggregate.Prices = pas
	r.reduced = true
	return r.aggregate.Clone()
}

func median(xs []uint64) uint64 {
	count := len(xs)
	if count == 0 {
		return 0
	}

	// Sort
	sort.Slice(xs, func(i, j int) bool { return xs[i] > xs[j] })

	if count%2 == 0 {
		// Even price point count, take the mean of the two middle prices
		i := int(count / 2)
		x1 := xs[i-1]
		x2 := xs[i]
		return uint64((x1 + x2) / 2)
	}
	// Odd price point count, use the middle price
	i := int((count - 1) / 2)
	return xs[i]
}

type IndirectMedian struct {
	pair   *model.Pair
	prices []*model.PriceAggregate
}

func NewIndirectMedian(pair *model.Pair) *IndirectMedian {
	return &IndirectMedian{pair: pair}
}

func (im *IndirectMedian) Ingest(pa *model.PriceAggregate) {
	if im.pair.Equal(pa.Pair) {
		im.prices = append(im.prices, pa)
	}
}

func (im *IndirectMedian) Aggregate(pair *model.Pair) *model.PriceAggregate {
	if !im.pair.Equal(pair) {
		return nil
	}

	var prices []uint64
	for _, pa := range im.prices {
		prices = append(prices, pa.Price)
	}

	return model.NewPriceAggregate(
		"indirect-median",
		&model.PricePoint{Pair: im.pair, Price: median(prices)},
		im.prices...,
	)
}
