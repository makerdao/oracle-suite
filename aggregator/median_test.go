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

	. "makerdao/gofer/model"
)

func TestOddPriceCount(t *testing.T) {
	rows := []*PriceAggregate{
		// Should be filtered due to outside time window
		newTestPricePointAggregate(-1000, "exchange0", "a", "b", 1000, 1),
		// Should be overwritten by entry 3 due to same exchange but older
		newTestPricePointAggregate(1, "exchange1", "a", "b", 2000, 1),
		newTestPricePointAggregate(2, "exchange2", "a", "b", 20, 1),
		newTestPricePointAggregate(3, "exchange1", "a", "b", 3, 1),
		// Should be skipped due to non-matching pair
		newTestPricePointAggregate(4, "exchange4", "n", "o", 4, 1),
		newTestPricePointAggregate(5, "exchange5", "a", "b", 5, 1),
	}

	for i := 0; i < 100; i++ {
		reducer := NewMedian(1000)
		pa := randomReduce(reducer, &Pair{Base: "a", Quote: "b"}, rows)
		assert.Equal(t, 3, len(pa.Prices), "length of aggregate price list")
		assert.Equal(t, uint64(5), pa.Price, "aggregate price should be median of price points")
	}
}

func TestEvenPriceCount(t *testing.T) {
	rows := []*PriceAggregate{
		newTestPricePointAggregate(1, "exchange1", "a", "b", 7, 1),
		newTestPricePointAggregate(2, "exchange2", "a", "b", 2, 1),
		newTestPricePointAggregate(3, "exchange3", "a", "b", 10, 1),
		newTestPricePointAggregate(4, "exchange4", "a", "b", 5, 1),
	}

	for i := 0; i < 100; i++ {
		reducer := NewMedian(1000)
		pa := randomReduce(reducer, &Pair{Base: "a", Quote: "b"}, rows)
		assert.Equal(t, 4, len(pa.Prices), "length of aggregate price list")
		assert.Equal(t, uint64(6), pa.Price, "aggregate price should be median of price points")
	}
}

func TestAskBidPriceFallback(t *testing.T) {
	rows := []*PriceAggregate{
		newTestPricePointAggregatePriceOnly(2, "exchange2", "a", "b", 2, 1),
		// No ask/bid and invalid last price
		newTestPricePointAggregatePriceOnly(1, "exchange1", "a", "b", 0, 1),
		newTestPricePointAggregate(4, "exchange4", "a", "b", 5, 1),
		// Invalid last price
		newTestPricePointAggregate(3, "exchange3", "a", "b", 0, 1),
	}

	for i := 0; i < 100; i++ {
		reducer := NewMedian(1000)
		pa := randomReduce(reducer, &Pair{Base: "a", Quote: "b"}, rows)
		assert.Equal(t, 2, len(pa.Prices), "length of aggregate price list")
		assert.Equal(t, uint64(3), pa.Price, "aggregate price should be median of price points")
	}
}

func TestInvalidPair(t *testing.T) {
	rows := []*PriceAggregate{
		newTestPricePointAggregatePriceOnly(1, "exchange1", "a", "b", 1, 1),
		newTestPricePointAggregate(2, "exchange4", "a", "b", 5, 1),
	}

	for i := 0; i < 100; i++ {
		reducer := NewMedian(1000)
		pa := randomReduce(reducer, &Pair{Base: "x", Quote: "y"}, rows)
		assert.Nil(t, pa)
	}
}

func TestIndirectMedian(t *testing.T) {
	// Only pair and price are considered in IndirectMedian
	pas := []*PriceAggregate{
		newTestPricePointAggregate(99999, "", "a", "b", 2, 1),
		newTestPricePointAggregate(3, "any exchange", "a", "b", 6, 1),
		newTestPricePointAggregatePriceOnly(1, "", "a", "b", 20, 1),
	}
	ignoredPAs := []*PriceAggregate{
		// Ignored, non matchin pair
		newTestPricePointAggregate(4, "", "x", "y", 1001, 1),
	}

	for i := 0; i < 100; i++ {
		reducer := NewIndirectMedian(&Pair{Base: "a", Quote: "b"})

		res := randomReduce(reducer, &Pair{Base: "a", Quote: "b"}, append(pas, ignoredPAs...))
		resFail := reducer.Aggregate(&Pair{Base: "x", Quote: "y"})

		assert.NotNil(t, res)
		assert.Equal(t, uint64(6), res.Price)
		assert.Equal(t, &Pair{Base: "a", Quote: "b"}, res.Pair)
		assert.ElementsMatch(t, pas, res.Prices)

		assert.Nil(t, resFail)
	}
}
