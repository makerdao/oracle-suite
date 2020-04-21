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

func NewTestPricePoint(timestamp int64, exchange string, base string, quote string, last uint64, volume uint64) *model.PricePoint {
	return &model.PricePoint{
		Timestamp: timestamp,
		Exchange:  &model.Exchange{Name: exchange},
		Pair:      &model.Pair{Base: base, Quote: quote},
		Last:      last,
		Ask:       last,
		Bid:       last,
		Volume:    volume,
	}
}

func NewTestPricePointPriceOnly(timestamp int64, exchange string, base string, quote string, last uint64, volume uint64) *model.PricePoint {
	return &model.PricePoint{
		Timestamp: timestamp,
		Exchange:  &model.Exchange{Name: exchange},
		Pair:      &model.Pair{Base: base, Quote: quote},
		Last:      last,
		Volume:    volume,
	}
}
