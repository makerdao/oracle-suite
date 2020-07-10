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

	"github.com/makerdao/gofer/model"
)

func TestTradeAggregator(t *testing.T) {
	pas := []*model.PriceAggregate{
		newTestPricePointAggregate(0, "exchange-a", "b", "a", 4, 1),
		newTestPricePointAggregate(0, "exchange-a", "b", "c", 8, 1),
		newTestPricePointAggregate(0, "exchange-a", "c", "d", 2, 1),
	}

	trade := NewTrade()
	res := trade.Aggregate(model.NewPair("a", "d"))
	assert.Nil(t, res)

	trade.Ingest(pas[0])
	res = trade.Aggregate(model.NewPair("b", "a"))
	assert.NotNil(t, res)
	assert.Equal(t, model.NewPair("b", "a"), res.Pair)

	trade.Ingest(pas[1])
	res = trade.Aggregate(model.NewPair("a", "c"))
	assert.NotNil(t, res)
	assert.Equal(t, model.NewPair("a", "c"), res.Pair)

	trade.Ingest(pas[2])

	res = trade.Aggregate(model.NewPair("a", "d"))

	assert.NotNil(t, res)
	assert.Equal(t, model.NewPair("a", "d"), res.Pair)
	assert.Equal(t, 4.0, res.Price)
	assert.ElementsMatch(t, pas, res.Prices)

	resFail := trade.Aggregate(model.NewPair("x", "y"))
	assert.Nil(t, resFail)
}
