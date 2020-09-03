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

package prcessor

import (
	"math/rand"

	"github.com/makerdao/gofer/pkg/gofer"
	"github.com/makerdao/gofer/pkg/model"
)

type Processor struct {
	ReturnsErr error
	Returns    []*model.PriceAggregate
	Pairs      []*model.Pair
}

func (mp *Processor) Process(agg gofer.IngestingAggregator, pairs ...*model.Pair) error {
	RandomReduce(agg, mp.Pairs, mp.Returns)
	return mp.ReturnsErr
}

func RandomReduce(r gofer.IngestingAggregator, pairs []*model.Pair, prices []*model.PriceAggregate) {
	for _, i := range rand.Perm(len(prices)) {
		r.Ingest(prices[i])
		if rand.Intn(2) == 1 {
			r.Aggregate(RandPair(pairs))
		}
	}
	r.Aggregate(RandPair(pairs))
}

func RandPair(pairs []*model.Pair) *model.Pair {
	pairsCount := len(pairs)
	if pairsCount == 0 {
		return nil
	}
	return pairs[rand.Intn(pairsCount)]
}
