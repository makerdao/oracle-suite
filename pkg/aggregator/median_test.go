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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/makerdao/gofer/pkg/model"
)

func TestOddPriceCount(t *testing.T) {
	rows := []*model.PriceAggregate{
		// nil is ignored
		nil,
		// Should be filtered due to outside time window
		newTestPricePointAggregate(-1000, "exchange0", "a", "b", 1000, 1),
		// Should be overwritten by entry 3 due to same exchange but older
		newTestPricePointAggregate(1, "exchange1", "a", "b", 2000, 1),
		newTestPricePointAggregate(2, "exchange2", "a", "b", 20, 1),
		newTestPricePointAggregate(3, "exchange1", "a", "b", 3, 1),
		// Should be skipped due to non-matching pair
		newTestPricePointAggregate(4, "exchange4", "n", "o", 1337, 1),
		newTestPricePointAggregate(5, "exchange5", "a", "b", 5, 1),
	}

	for i := 0; i < 100; i++ {
		reducer := NewMedian(nil, 1000)
		pa := randomReduce(reducer, model.NewPair("a", "b"), rows)
		assert.Nil(t, reducer.Aggregate(nil))
		assert.Equal(t, 3, len(pa.Prices), "length of aggregate price list")
		assert.Equal(t, 5.0, pa.Price, "aggregate price should be median of price points")

		paNO := randomReduce(reducer, model.NewPair("n", "o"), rows)
		assert.Equal(t, 1, len(paNO.Prices), "length of aggregate price list")
		assert.Equal(t, 1337.0, paNO.Price, "aggregate price should be median of price points")
	}
}

func TestEvenPriceCount(t *testing.T) {
	rows := []*model.PriceAggregate{
		newTestPricePointAggregate(1, "exchange1", "a", "b", 7, 1),
		newTestPricePointAggregate(2, "exchange2", "a", "b", 2, 1),
		newTestPricePointAggregate(3, "exchange3", "a", "b", 10, 1),
		newTestPricePointAggregate(4, "exchange4", "a", "b", 5, 1),
	}

	for i := 0; i < 100; i++ {
		reducer := NewMedian(nil, 1000)
		pa := randomReduce(reducer, model.NewPair("a", "b"), rows)
		assert.Equal(t, 4, len(pa.Prices), "length of aggregate price list")
		assert.Equal(t, 6.0, pa.Price, "aggregate price should be median of price points")
	}
}

func TestAskBidPriceFallback(t *testing.T) {
	rows := []*model.PriceAggregate{
		newTestPricePointAggregatePriceOnly(2, "exchange2", "a", "b", 2, 1),
		// No ask/bid and invalid last price
		newTestPricePointAggregatePriceOnly(1, "exchange1", "a", "b", 0, 1),
		newTestPricePointAggregate(4, "exchange4", "a", "b", 5, 1),
		// Invalid last price
		newTestPricePointAggregate(3, "exchange3", "a", "b", 0, 1),
	}

	for i := 0; i < 100; i++ {
		reducer := NewMedian(nil, 1000)
		pa := randomReduce(reducer, model.NewPair("a", "b"), rows)
		assert.Equal(t, 2, len(pa.Prices), "length of aggregate price list")
		assert.Equal(t, 3.5, pa.Price, "aggregate price should be median of price points")
	}
}

func TestInvalidPair(t *testing.T) {
	rows := []*model.PriceAggregate{
		newTestPricePointAggregatePriceOnly(1, "exchange1", "a", "b", 1, 1),
		newTestPricePointAggregate(2, "exchange4", "a", "b", 5, 1),
	}

	for i := 0; i < 100; i++ {
		reducer := NewMedian(nil, 1000)
		pa := randomReduce(reducer, model.NewPair("x", "y"), rows)
		assert.Nil(t, pa)
	}
}

func TestMedian(t *testing.T) {
	var res float64

	res = median([]float64{})
	assert.Equal(t, 0.0, res)

	res = median([]float64{4,2,3,4,5})
	assert.Equal(t, 4.0, res)

	res = median([]float64{5,2,10,19})
	assert.Equal(t, 7.5, res)
}
