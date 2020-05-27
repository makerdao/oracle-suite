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
	. "github.com/makerdao/gofer/aggregator"
	. "github.com/makerdao/gofer/model"
	"math/rand"
)

type mockProcessor struct {
	returnsErr error
	returns    []*PriceAggregate
	pairs      []*Pair
}

func (mp *mockProcessor) Process(ppps []*PotentialPricePoint, agg Aggregator) (Aggregator, error) {
	randomReduce(agg, mp.pairs, mp.returns)
	return agg, mp.returnsErr
}

func newTestPricePointAggregate(timestamp int64, exchange string, base string, quote string, price float64, volume float64) *PriceAggregate {
	return &PriceAggregate{
		PricePoint: &PricePoint{
			Timestamp: timestamp,
			Exchange:  &Exchange{Name: exchange},
			Pair:      &Pair{Base: base, Quote: quote},
			Price:     price,
			Ask:       price,
			Bid:       price,
			Volume:    volume,
		},
	}
}

func randPair(pairs []*Pair) *Pair {
  pairsCount := len(pairs)
  if pairsCount == 0 {
    return nil
  }
  return pairs[rand.Intn(pairsCount)]
}

func randomReduce(r Aggregator, pairs []*Pair, prices []*PriceAggregate) {
	for _, i := range rand.Perm(len(prices)) {
		r.Ingest(prices[i])
		if rand.Intn(2) == 1 {
			r.Aggregate(randPair(pairs))
		}
	}
	r.Aggregate(randPair(pairs))
}
