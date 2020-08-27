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

	mock "github.com/makerdao/gofer/internal/pkg/mock/aggregator"
	"github.com/makerdao/gofer/pkg/model"
)

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
	directAggregator := &mock.Aggregator{
		Returns: map[model.Pair]*model.PriceAggregate{
			*model.NewPair("a", "b"): abReturn,
			*model.NewPair("b", "d"): bdReturn,
			*model.NewPair("a", "c"): acReturn,
			*model.NewPair("c", "d"): cdReturn,
			*model.NewPair("b", "a"): baReturn,
		},
	}
	pas := []*model.PriceAggregate{
		newTestPricePointAggregate(0, "exchange3", "a", "b", 101, 1),
		newTestPricePointAggregate(0, "exchange4", "a", "c", 102, 1),
		newTestPricePointAggregate(0, "exchange5", "b", "d", 103, 1),
		newTestPricePointAggregate(0, "exchange5", "c", "d", 104, 1),
		newTestPricePointAggregate(0, "exchange5", "b", "a", 105, 1),
	}
	ppaths := []*model.PricePath{
		&model.PricePath{
			model.NewPair("a", "b"),
			model.NewPair("b", "d"),
		},
		&model.PricePath{
			model.NewPair("a", "c"),
			model.NewPair("c", "d"),
		},
		&model.PricePath{
			model.NewPair("b", "a"),
			model.NewPair("b", "d"),
		},
	}

	for i := 0; i < 100; i++ {
		pathAggregator := NewPathWithPathMap(
			ppaths,
			nil,
			directAggregator,
		)

		res := pathAggregator.Aggregate(nil)
		assert.Nil(t, res)

		res = pathAggregator.Aggregate(model.NewPair("x", "y"))
		assert.Nil(t, res)

		res = randomReduce(pathAggregator, model.NewPair("a", "d"), pas)
		assert.NotNil(t, res)
		assert.Equal(t, model.NewPair("a", "d"), res.Pair)
		assert.Equal(t, "path", res.PriceModelName)
		// Median of trade abd, acd and bad is 1001 * 1002
		assert.Equal(t, 1001.0 * 1002.0, res.Price)

		resTradeABD := res.Prices[0]
		assert.NotNil(t, resTradeABD)
		assert.Equal(t, model.NewPair("a", "d"), resTradeABD.Pair)
		assert.Equal(t, "trade", resTradeABD.PriceModelName)
		assert.Equal(t, 1001.0 * 1002.0, resTradeABD.Price)

		resMedinaAB := resTradeABD.Prices[0]
		assert.NotNil(t, resMedinaAB)
		assert.Equal(t, model.NewPair("a", "b"), resMedinaAB.Pair)
		assert.Equal(t, "median", resMedinaAB.PriceModelName)
		assert.Equal(t, 1001.0, resMedinaAB.Price)
		assert.Equal(t, abReturn, resMedinaAB)

		resMedinaBD := resTradeABD.Prices[1]
		assert.NotNil(t, resMedinaBD)
		assert.Equal(t, model.NewPair("b", "d"), resMedinaBD.Pair)
		assert.Equal(t, "median", resMedinaBD.PriceModelName)
		assert.Equal(t, 1002.0, resMedinaBD.Price)
		assert.Equal(t, bdReturn, resMedinaBD)

		resTradeACD := res.Prices[1]
		assert.NotNil(t, resTradeACD)
		assert.Equal(t, model.NewPair("a", "d"), resTradeACD.Pair)
		assert.Equal(t, "trade", resTradeACD.PriceModelName)
		assert.Equal(t, 1003.0 * 1004.0, resTradeACD.Price)

		resMedinaAC := resTradeACD.Prices[0]
		assert.NotNil(t, resMedinaAC)
		assert.Equal(t, model.NewPair("a", "c"), resMedinaAC.Pair)
		assert.Equal(t, "median", resMedinaAC.PriceModelName)
		assert.Equal(t, 1003.0, resMedinaAC.Price)
		assert.Equal(t, acReturn, resMedinaAC)

		resMedinaCD := resTradeACD.Prices[1]
		assert.NotNil(t, resMedinaCD)
		assert.Equal(t, model.NewPair("c", "d"), resMedinaCD.Pair)
		assert.Equal(t, "median", resMedinaCD.PriceModelName)
		assert.Equal(t, 1004.0, resMedinaCD.Price)
		assert.Equal(t, cdReturn, resMedinaCD)

		resTradeBAD := res.Prices[2]
		assert.NotNil(t, resTradeBAD)
		assert.Equal(t, model.NewPair("a", "d"), resTradeBAD.Pair)
		assert.Equal(t, "trade", resTradeBAD.PriceModelName)
		assert.Equal(t, 1002.0 / 2, resTradeBAD.Price)

		resMedinaBA := resTradeBAD.Prices[0]
		assert.NotNil(t, resMedinaBA)
		assert.Equal(t, model.NewPair("b", "a"), resMedinaBA.Pair)
		assert.Equal(t, "median", resMedinaAC.PriceModelName)
		assert.Equal(t, 2.0, resMedinaBA.Price)
		assert.Equal(t, baReturn, resMedinaBA)

		assert.Equal(t, resMedinaBD, resTradeBAD.Prices[1])
	}
}

func TestTrade(t *testing.T) {
	pasAE := []*model.PriceAggregate{
		newTestPricePointAggregate(0, "exchange1", "a", "b", 10, 1),
		newTestPricePointAggregate(0, "exchange1", "b", "c", 20, 1),
		newTestPricePointAggregate(0, "exchange1", "c", "d", 200, 1),
		newTestPricePointAggregate(0, "exchange1", "d", "e", 40, 1),
	}
	pasCE := []*model.PriceAggregate{
		newTestPricePointAggregate(0, "exchange1", "a", "b", 10, 1),
		newTestPricePointAggregate(0, "exchange1", "b", "c", 20, 1),
		newTestPricePointAggregate(0, "exchange1", "a", "d", 200, 1),
		newTestPricePointAggregate(0, "exchange1", "d", "e", 40, 1),
	}

	resAE := trade(pasAE)
	assert.NotNil(t, resAE)
	assert.Equal(t, model.NewPair("a", "e"), resAE.Pair)
	assert.Equal(t, 10.0 * 20.0 * 200.0 * 40.0, resAE.Price)

	resCE := trade(pasCE)
	assert.NotNil(t, resCE)
	assert.Equal(t, model.NewPair("c", "e"), resCE.Pair)
	assert.Equal(t, 200.0 / (10.0 * 20.0) * 40.0, resCE.Price)
}

func TestPathResolveMissingPair(t *testing.T) {
	ppath := &model.PricePath{
		model.NewPair("a", "b"),
		model.NewPair("b", "d"),
	}

	directAggregator := &mock.Aggregator{
		Returns: map[model.Pair]*model.PriceAggregate{
			*model.NewPair("a", "b"): newTestPricePointAggregate(0, "exchange5", "a", "b", 100, 1),
		},
	}

	pathAggregator := NewPathWithPathMap(
		[]*model.PricePath{ppath},
		nil,
		directAggregator,
	)

	var res *model.PriceAggregate

	res = pathAggregator.resolve(model.PricePath{ model.NewPair("a", "b") })
	assert.NotNil(t, res)
	assert.Equal(t, 100.0, res.Price)

	res = pathAggregator.resolve(*ppath)
	assert.Nil(t, res)
}
