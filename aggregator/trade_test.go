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

func TestTradeAggregator(t *testing.T) {
	pas := []*PriceAggregate{
		newTestPricePointAggregate(0, "exchange-a", "b", "a", 4, 1),
		newTestPricePointAggregate(0, "exchange-a", "b", "c", 8, 1),
		newTestPricePointAggregate(0, "exchange-a", "c", "d", 2, 1),
	}

	trade := NewTrade()
	res := trade.Aggregate(&Pair{Base: "a", Quote: "d"})
	assert.Nil(t, res)

	trade.Ingest(pas[0])
	res = trade.Aggregate(&Pair{Base: "b", Quote: "a"})
	assert.NotNil(t, res)
	assert.Equal(t, &Pair{Base: "b", Quote: "a"}, res.Pair)

	trade.Ingest(pas[1])
	res = trade.Aggregate(&Pair{Base: "a", Quote: "c"})
	assert.NotNil(t, res)
	assert.Equal(t, &Pair{Base: "a", Quote: "c"}, res.Pair)

	trade.Ingest(pas[2])

	res = trade.Aggregate(&Pair{Base: "a", Quote: "d"})

	assert.NotNil(t, res)
	assert.Equal(t, &Pair{Base: "a", Quote: "d"}, res.Pair)
	assert.Equal(t, uint64(4), res.Price)
	assert.ElementsMatch(t, pas, res.Prices)

	resFail := trade.Aggregate(&Pair{Base: "x", Quote: "y"})
	assert.Nil(t, resFail)
}
