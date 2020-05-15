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

type mockAggregator struct {
	returns map[Pair]*PriceAggregate
}

func (mr *mockAggregator) Ingest(pa *PriceAggregate) {
}

func (mr *mockAggregator) Aggregate(pair *Pair) *PriceAggregate {
	return mr.returns[*pair]
}

func TestPathAggregator(t *testing.T) {
	abReturn := newTestPriceAggregate("median", "a", "b", 1001,
		newTestPricePointAggregate(0, "exchange1", "a", "b", 1, 1),
		newTestPricePointAggregate(0, "exchange2", "a", "b", 2, 1),
	)
	bdReturn := newTestPriceAggregate("median", "b", "d", 1002,
		newTestPricePointAggregate(0, "exchange3", "b", "d", 3, 1),
		newTestPricePointAggregate(0, "exchange4", "b", "d", 4, 1),
		newTestPricePointAggregate(0, "exchange5", "b", "d", 5, 1),
	)
	acReturn := newTestPriceAggregate("median", "a", "c", 1003,
		newTestPricePointAggregate(0, "exchange1", "a", "c", 6, 1),
		newTestPricePointAggregate(0, "exchange2", "a", "c", 7, 1),
		newTestPricePointAggregate(0, "exchange3", "a", "c", 8, 1),
	)
	cdReturn := newTestPriceAggregate("median", "c", "d", 1004,
		newTestPricePointAggregate(0, "exchange3", "c", "d", 9, 1),
	)
	baReturn := newTestPriceAggregate("median", "b", "a", 2,
		newTestPricePointAggregate(0, "exchange1", "b", "a", 10, 1),
	)
	directAggregator := &mockAggregator{
		returns: map[Pair]*PriceAggregate{
			Pair{Base: "a", Quote: "b"}: abReturn,
			Pair{Base: "b", Quote: "d"}: bdReturn,
			Pair{Base: "a", Quote: "c"}: acReturn,
			Pair{Base: "c", Quote: "d"}: cdReturn,
			Pair{Base: "b", Quote: "a"}: baReturn,
		},
	}
	pas := []*PriceAggregate{
		newTestPricePointAggregate(0, "exchange3", "a", "b", 101, 1),
		newTestPricePointAggregate(0, "exchange4", "a", "c", 102, 1),
		newTestPricePointAggregate(0, "exchange5", "b", "d", 103, 1),
		newTestPricePointAggregate(0, "exchange5", "c", "d", 104, 1),
		newTestPricePointAggregate(0, "exchange5", "b", "a", 105, 1),
	}
	ppathss := []*PricePaths{
		NewPricePaths(
			NewPair("a", "d"),
			[]*Pair{
				NewPair("a", "b"),
				NewPair("b", "d"),
			},
			[]*Pair{
				NewPair("a", "c"),
				NewPair("c", "d"),
			},
			[]*Pair{
				NewPair("b", "a"),
				NewPair("b", "d"),
			},
		),
	}

	for i := 0; i < 100; i++ {
		pathAggregator := NewPath(
			ppathss,
			directAggregator,
		)

		res := randomReduce(pathAggregator, NewPair("a", "d"), pas)
		assert.NotNil(t, res)
		assert.Equal(t, &Pair{Base: "a", Quote: "d"}, res.Pair)
		assert.Equal(t, "indirect-median", res.PriceModelName)
		assert.Equal(t, uint64(1001 * 1002), res.Price)

		resTradeABD := res.Prices[0]
		assert.NotNil(t, resTradeABD)
		assert.Equal(t, &Pair{Base: "a", Quote: "d"}, resTradeABD.Pair)
		assert.Equal(t, "trade", resTradeABD.PriceModelName)
		assert.Equal(t, uint64(1001 * 1002), resTradeABD.Price)

		resMedinaAB := resTradeABD.Prices[0]
		assert.NotNil(t, resMedinaAB)
		assert.Equal(t, &Pair{Base: "a", Quote: "b"}, resMedinaAB.Pair)
		assert.Equal(t, "median", resMedinaAB.PriceModelName)
		assert.Equal(t, uint64(1001), resMedinaAB.Price)
		assert.Equal(t, abReturn, resMedinaAB)

		resMedinaBD := resTradeABD.Prices[1]
		assert.NotNil(t, resMedinaBD)
		assert.Equal(t, &Pair{Base: "b", Quote: "d"}, resMedinaBD.Pair)
		assert.Equal(t, "median", resMedinaBD.PriceModelName)
		assert.Equal(t, uint64(1002), resMedinaBD.Price)
		assert.Equal(t, bdReturn, resMedinaBD)

		resTradeACD := res.Prices[1]
		assert.NotNil(t, resTradeACD)
		assert.Equal(t, &Pair{Base: "a", Quote: "d"}, resTradeACD.Pair)
		assert.Equal(t, "trade", resTradeACD.PriceModelName)
		assert.Equal(t, uint64(1003 * 1004), resTradeACD.Price)

		resMedinaAC := resTradeACD.Prices[0]
		assert.NotNil(t, resMedinaAC)
		assert.Equal(t, &Pair{Base: "a", Quote: "c"}, resMedinaAC.Pair)
		assert.Equal(t, "median", resMedinaAC.PriceModelName)
		assert.Equal(t, uint64(1003), resMedinaAC.Price)
		assert.Equal(t, acReturn, resMedinaAC)

		resMedinaCD := resTradeACD.Prices[1]
		assert.NotNil(t, resMedinaCD)
		assert.Equal(t, &Pair{Base: "c", Quote: "d"}, resMedinaCD.Pair)
		assert.Equal(t, "median", resMedinaCD.PriceModelName)
		assert.Equal(t, uint64(1004), resMedinaCD.Price)
		assert.Equal(t, cdReturn, resMedinaCD)

		resTradeBAD := res.Prices[2]
		assert.NotNil(t, resTradeBAD)
		assert.Equal(t, &Pair{Base: "a", Quote: "d"}, resTradeBAD.Pair)
		assert.Equal(t, "trade", resTradeBAD.PriceModelName)
		assert.Equal(t, uint64(1002 / 2), resTradeBAD.Price)

		resMedinaBA := resTradeBAD.Prices[0]
		assert.NotNil(t, resMedinaBA)
		assert.Equal(t, &Pair{Base: "b", Quote: "a"}, resMedinaBA.Pair)
		assert.Equal(t, "median", resMedinaAC.PriceModelName)
		assert.Equal(t, uint64(2), resMedinaBA.Price)
		assert.Equal(t, baReturn, resMedinaBA)

		assert.Equal(t, resMedinaBD, resTradeBAD.Prices[1])
	}
}

func TestTrade(t *testing.T) {
	pasAE := []*PriceAggregate{
		newTestPricePointAggregate(0, "exchange1", "a", "b", 10, 1),
		newTestPricePointAggregate(0, "exchange1", "b", "c", 20, 1),
		newTestPricePointAggregate(0, "exchange1", "c", "d", 200, 1),
		newTestPricePointAggregate(0, "exchange1", "d", "e", 40, 1),
	}
	pasCE := []*PriceAggregate{
		newTestPricePointAggregate(0, "exchange1", "a", "b", 10, 1),
		newTestPricePointAggregate(0, "exchange1", "b", "c", 20, 1),
		newTestPricePointAggregate(0, "exchange1", "a", "d", 200, 1),
		newTestPricePointAggregate(0, "exchange1", "d", "e", 40, 1),
	}

	resAE := trade(pasAE)
	assert.NotNil(t, resAE)
	assert.Equal(t, &Pair{Base: "a", Quote: "e"}, resAE.Pair)
	assert.Equal(t, uint64(1600000), resAE.Price)

	resCE := trade(pasCE)
	assert.NotNil(t, resCE)
	assert.Equal(t, &Pair{Base: "c", Quote: "e"}, resCE.Pair)
	assert.Equal(t, uint64(40), resCE.Price)
}
