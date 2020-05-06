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
		newTestPricePointAggregate(0, "exchange-a", "a", "b", 2, 1),
		newTestPricePointAggregate(0, "exchange-a", "b", "c", 4, 1),
		newTestPricePointAggregate(0, "exchange-a", "c", "d", 1, 1),
	}

	trade := NewTrade(&Pair{Base: "a", Quote: "d"})

	for _, pa := range pas {
		trade.Ingest(pa)
	}

	res := trade.Aggregate(&Pair{Base: "a", Quote: "d"})
	resFail := trade.Aggregate(&Pair{Base: "x", Quote: "y"})

	assert.NotNil(t, res)
	assert.Equal(t, uint64(8), res.Price)
	assert.ElementsMatch(t, pas, res.Prices)
	assert.Nil(t, resFail)
}
