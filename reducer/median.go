package reducer

import "makerdao/gofer/model"

type MedianReducer struct {
	timeWindow int64
}

func NewMedianReducer(timeWindow int64) *MedianReducer {
	return &MedianReducer{
		timeWindow: timeWindow,
	}
}

func (r *MedianReducer) Reduce(aggregate *model.PriceAggregate, price *model.PricePoint) *model.PriceAggregate {
	var newest *model.PricePoint
	if price.Base != aggregate.Base || price.Quote != aggregate.Quote || price.Price == 0 {
		return aggregate
	}

	// Add new price and find newest
	updatedPrices := []*model.PricePoint{price}
	newest = price
	for _, p := range aggregate.Prices {
		if p.Timestamp > newest.Timestamp {
			newest = p
		}
		if price.Exchange == p.Exchange {
			if price.Timestamp > p.Timestamp {
				continue
			} else {
				return aggregate
			}
		}
		updatedPrices = append(updatedPrices, p)
	}

	// Filter inside time window
	timewindowedPrices := []*model.PricePoint{}
	for _, p := range updatedPrices {
		if p.Timestamp > (newest.Timestamp - r.timeWindow) {
			timewindowedPrices = append(timewindowedPrices, p)
		}
	}
	aggregate.NewestTimestamp = newest.Timestamp
	aggregate.Prices = timewindowedPrices

	priceCount := len(aggregate.Prices)
	if priceCount == 0 {
		return aggregate
	}

	if priceCount%2 == 0 {
		i := int(priceCount / 2)
		aggregate.Price = uint64((aggregate.Prices[i-1].Price + aggregate.Prices[i].Price) / 2)
	} else {
		i := int((priceCount - 1) / 2)
		aggregate.Price = aggregate.Prices[i].Price
	}

	return aggregate
}
