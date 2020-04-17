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

// Get price estimate from price point
func calcPrice(pp *model.PricePoint) uint64 {
	// If ask/bid values are valid return mean of ask and bid
	if pp.Ask != 0 && pp.Bid != 0 {
		return (pp.Ask + pp.Bid) / 2
	}
	// If last auction price is valid return that
	if pp.Last != 0 {
		return pp.Last
	}
	// Otherwise return invalid price
	return 0
}

func NewMedianReducer(pair *model.Pair, timeWindow int64) *MedianReducer {
	return &MedianReducer{
		timeWindow:      timeWindow,
		newestTimestamp: 0,
		aggregate:       model.NewPriceAggregate(pair),
		reduced:         false,
	}
}

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

	// First ingested price, add it and return
	if len(r.aggregate.Prices) == 0 {
		r.newestTimestamp = pp.Timestamp
		r.aggregate.Prices = []*model.PricePoint{pp}
		r.reduced = false
		return
	}

	if pp.Timestamp > r.newestTimestamp {
		r.newestTimestamp = pp.Timestamp
	}

	timeWindow := r.newestTimestamp - r.timeWindow
	// New price is outside time window, do nothing
	if pp.Timestamp <= timeWindow {
		return
	}

	var updatedIgnested []*model.PricePoint
	addPrice := true
	for _, p := range r.aggregate.Prices {
		// Remove prices outside time window
		if p.Timestamp <= timeWindow {
			continue
		}
		if pp.Exchange == p.Exchange {
			// Price with same exchange as new price already exists
			if pp.Timestamp > p.Timestamp {
				// Update existing price if new price is newer
				p.Ask = pp.Ask
				p.Bid = pp.Bid
				p.Last = pp.Last
				p.Volume = pp.Volume
				p.Timestamp = pp.Timestamp
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
		updatedIgnested = append(updatedIgnested, pp)
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
		return calcPrice(r.aggregate.Prices[i]) > calcPrice(r.aggregate.Prices[j])
	})

	if priceCount%2 == 0 {
		// Even price point count, take the mean of the two middle prices
		i := int(priceCount / 2)
		price1 := calcPrice(r.aggregate.Prices[i-1])
		price2 := calcPrice(r.aggregate.Prices[i])
		r.aggregate.Price = uint64((price1 + price2) / 2)
	} else {
		// Odd price point count, use the middle price
		i := int((priceCount - 1) / 2)
		r.aggregate.Price = calcPrice(r.aggregate.Prices[i])
	}
	r.reduced = true
	return r.aggregate.Clone()
}
