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
