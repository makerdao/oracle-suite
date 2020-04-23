//  Copyright (C) 2020  Maker Foundation
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

package reducer

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"makerdao/gofer/model"
)

func TestOddPriceCount(t *testing.T) {
	rows := []*model.PricePoint{
		// Should be filtered due to outside time window
		NewTestPricePoint(-1000, "exchange0", "a", "b", 1000, 1),
		// Should be overwritten by entry 3 due to same exchange but older
		NewTestPricePoint(1, "exchange1", "a", "b", 2000, 1),
		NewTestPricePoint(2, "exchange2", "a", "b", 20, 1),
		NewTestPricePoint(3, "exchange1", "a", "b", 3, 1),
		// Should be skipped due to non-matching pair
		NewTestPricePoint(4, "exchange4", "n", "o", 4, 1),
		NewTestPricePoint(5, "exchange5", "a", "b", 5, 1),
	}

	for i := 0; i < 100; i++ {
		reducer := NewMedianReducer(&model.Pair{Base: "a", Quote: "b"}, 1000)
		priceAggregate := RandomReduce(reducer, rows)
		assert.Equal(t, 3, len(priceAggregate.Prices), "length of aggregate price list")
		assert.Equal(t, uint64(5), priceAggregate.Price, "aggregate price should be median of price points")
	}
}

func TestEvenPriceCount(t *testing.T) {
	rows := []*model.PricePoint{
		NewTestPricePoint(1, "exchange1", "a", "b", 7, 1),
		NewTestPricePoint(2, "exchange2", "a", "b", 2, 1),
		NewTestPricePoint(3, "exchange3", "a", "b", 10, 1),
		NewTestPricePoint(4, "exchange4", "a", "b", 5, 1),
	}

	for i := 0; i < 100; i++ {
		reducer := NewMedianReducer(&model.Pair{Base: "a", Quote: "b"}, 1000)
		priceAggregate := RandomReduce(reducer, rows)
		assert.Equal(t, 4, len(priceAggregate.Prices), "length of aggregate price list")
		assert.Equal(t, uint64(6), priceAggregate.Price, "aggregate price should be median of price points")
	}
}

func TestAskBidPriceFallback(t *testing.T) {
	rows := []*model.PricePoint{
		NewTestPricePointPriceOnly(2, "exchange2", "a", "b", 2, 1),
		// No ask/bid and invalid last price
		NewTestPricePointPriceOnly(1, "exchange1", "a", "b", 0, 1),
		NewTestPricePoint(4, "exchange4", "a", "b", 5, 1),
		// Invalid last price
		NewTestPricePoint(3, "exchange3", "a", "b", 0, 1),
	}

	for i := 0; i < 100; i++ {
		reducer := NewMedianReducer(&model.Pair{Base: "a", Quote: "b"}, 1000)
		priceAggregate := RandomReduce(reducer, rows)
		assert.Equal(t, 2, len(priceAggregate.Prices), "length of aggregate price list")
		assert.Equal(t, uint64(3), priceAggregate.Price, "aggregate price should be median of price points")
	}
}
