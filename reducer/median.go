package reducer

import (
	"sort"

	"makerdao/gofer/model"
)

type MedianReducer struct {
	timeWindow      int64
	aggregate       *model.PriceAggregate
	newestTimestamp int64
	reduced         bool
}

func NewMedianReducer(pair *model.Pair, timeWindow int64) *MedianReducer {
	return &MedianReducer{
		timeWindow:      timeWindow,
		newestTimestamp: 0,
		aggregate:       model.NewPriceAggregate(pair),
		reduced:         false,
	}
}

func (r *MedianReducer) Ingest(price *model.PricePoint) {
	if price.Price == 0 || !price.Pair.Equal(r.aggregate.Pair) {
		return
	}

	// First ingested price, add it and return
	if len(r.aggregate.Prices) == 0 {
		r.newestTimestamp = price.Timestamp
		r.aggregate.Prices = []*model.PricePoint{price}
		r.reduced = false
		return
	}

	if price.Timestamp > r.newestTimestamp {
		r.newestTimestamp = price.Timestamp
	}

	timeWindow := r.newestTimestamp - r.timeWindow
	// New price is outside time window, do nothing
	if price.Timestamp <= timeWindow {
		return
	}

	var updatedIgnested []*model.PricePoint
	addPrice := true
	for _, p := range r.aggregate.Prices {
		// Remove prices outside time window
		if p.Timestamp <= timeWindow {
			continue
		}
		if price.Exchange == p.Exchange {
			// Price with same exchange as new price already exists
			if price.Timestamp > p.Timestamp {
				// Update existing price if new price is newer
				p.Price = price.Price
				p.Volume = price.Volume
				p.Timestamp = price.Timestamp
				addPrice = false
			} else {
				// New price is older than existing price, do nothing
				return
			}
		}
		updatedIgnested = append(updatedIgnested, p)
	}

	// Add new price if not already updated existing price with same exchange
	if addPrice {
		updatedIgnested = append(updatedIgnested, price)
	}

	r.aggregate.Prices = updatedIgnested
	r.reduced = false
}

func (r *MedianReducer) Reduce() *model.PriceAggregate {
	priceCount := len(r.aggregate.Prices)
	if priceCount == 0 || r.reduced {
		return r.aggregate.Clone()
	}

	// Sort price points by price
	sort.Slice(r.aggregate.Prices, func(i, j int) bool {
		return r.aggregate.Prices[i].Price > r.aggregate.Prices[j].Price
	})

	if priceCount%2 == 0 {
		// Even price point count, take the mean of the two middle prices
		i := int(priceCount / 2)
		r.aggregate.Price = uint64((r.aggregate.Prices[i-1].Price + r.aggregate.Prices[i].Price) / 2)
	} else {
		// Odd price point count, use the middle price
		i := int((priceCount - 1) / 2)
		r.aggregate.Price = r.aggregate.Prices[i].Price
	}
	r.reduced = true
	return r.aggregate.Clone()
}
