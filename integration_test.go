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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/makerdao/gofer/aggregator"
	"github.com/makerdao/gofer/mock"
	"github.com/makerdao/gofer/model"
)

func TestPathWithSetzerPatherAndMedianIntegration(t *testing.T) {
	pairs := []*model.Pair{
		{Base: "ETH", Quote: "USD"},
		{Base: "ETH", Quote: "BTC"},
		{Base: "BTC", Quote: "USD"},
		{Base: "REP", Quote: "USD"},
		{Base: "USDC", Quote: "USD"},
	}

	pas := []*model.PriceAggregate{
		newTestPricePointAggregate(6, "exchange1", "ETH", "BTC", 2, 1),
		newTestPricePointAggregate(7, "exchange2", "ETH", "BTC", 4, 1),
		// median( ()=>ETH/BTC$2, ()=>ETH/BTC$4 )=>ETH/BTC$3

		// Should be filtered due to outside time window
		newTestPricePointAggregate(-1000, "exchange0", "BTC", "USD", 1000, 1),
		// Should be overwritten by entry 3 due to same exchange but older
		newTestPricePointAggregate(1, "exchange1", "BTC", "USD", 2000, 1),
		newTestPricePointAggregate(2, "exchange2", "BTC", "USD", 20, 1),
		newTestPricePointAggregate(3, "exchange1", "BTC", "USD", 3, 1),
		// Should be skipped due to non-matching pair
		newTestPricePointAggregate(4, "exchange4", "n", "o", 4, 1),
		newTestPricePointAggregate(5, "exchange5", "BTC", "USD", 10, 1),
		// median( ()=>BTC/USD$3, ()=>BTC/USD$10, ()=>BTC/USD$20 )=>BTC/USD$10
		// trade( median(..)=>ETH/BTC$3, median(..)=>BTC/USD$10 )=>ETH/USD$30

		newTestPricePointAggregate(8, "exchange1", "ETH", "USDT", 3, 1),
		newTestPricePointAggregate(9, "exchange2", "USDT", "USD", 3, 1),
		// trade( median(..)=>ETH/USDT$3, medain(..)=>USDT/USD$3 )=>ETH/USD$9

		newTestPricePointAggregate(10, "exchange1", "BTC", "USDC", 2, 1),
		// median( ()=>BTC/USDC$2 )=>BTC/USDC$2
		// trade( median(..)=>BTC/USDC$2, medain(..)=>BTC/USD$10 )=>USDC/USD$5
	}
	// path( trade(..)=>ETH/USD$9, trade(..)=>ETH/USD$30 )=>ETH/USD$19
	// path( trade(..)=>USDC/USD$5 )=>USDC/USD$5

	sources := []*model.PotentialPricePoint{}

	agg := aggregator.NewPath(aggregator.NewSetzer(), sources, aggregator.NewMedian(nil, 1000))

	processor := &mock.Processor{
		Returns: pas,
		Pairs:   append(pairs, nil),
	}

	gofer := NewGofer(agg, processor)

	for i := 0; i < 100; i++ {
		res, err := gofer.Prices(pairs...)

		assert.NoError(t, err)

		res_ETH_USD := res[model.Pair{Base: "ETH", Quote: "USD"}]
		assert.NotNil(t, res_ETH_USD)
		assert.Equal(t, &model.Pair{Base: "ETH", Quote: "USD"}, res_ETH_USD.Pair)
		assert.Equal(t, "path", res_ETH_USD.PriceModelName)
		assert.Equal(t, 19.5, res_ETH_USD.Price)

		res_ETH_BTC := res[model.Pair{Base: "ETH", Quote: "BTC"}]
		assert.NotNil(t, res_ETH_BTC)
		assert.Equal(t, &model.Pair{Base: "ETH", Quote: "BTC"}, res_ETH_BTC.Pair)
		assert.Equal(t, "path", res_ETH_BTC.PriceModelName)
		assert.Equal(t, 3.0, res_ETH_BTC.Price)

		res_BTC_USD := res[model.Pair{Base: "BTC", Quote: "USD"}]
		assert.NotNil(t, res_BTC_USD)
		assert.Equal(t, &model.Pair{Base: "BTC", Quote: "USD"}, res_BTC_USD.Pair)
		assert.Equal(t, "path", res_BTC_USD.PriceModelName)
		assert.Equal(t, 10.0, res_BTC_USD.Price)

		res_ETH_KRW := res[model.Pair{Base: "ETH", Quote: "KRW"}]
		assert.Nil(t, res_ETH_KRW, "Pair not existing in Pather")

		res_REP_USD := res[model.Pair{Base: "REP", Quote: "USD"}]
		assert.NotNil(t, res_REP_USD, "Pair existis in Pather but no price points yet")
		assert.Equal(t, &model.Pair{Base: "REP", Quote: "USD"}, res_REP_USD.Pair)
		assert.Equal(t, "path", res_REP_USD.PriceModelName)
		assert.Equal(t, 0.0, res_REP_USD.Price)

		res_USDC_USD := res[model.Pair{Base: "USDC", Quote: "USD"}]
		assert.NotNil(t, res_USDC_USD)
		assert.Equal(t, &model.Pair{Base: "USDC", Quote: "USD"}, res_USDC_USD.Pair)
		assert.Equal(t, "path", res_USDC_USD.PriceModelName)
		assert.Equal(t, 5.0, res_USDC_USD.Price)
	}
}

func TestSetzAggregatorIntegration(t *testing.T) {
	gofer, err := ReadFile("testdata/integration-config-1.json")

	assert.NoError(t, err)
	assert.NotNil(t, gofer)

	pairs := []*model.Pair{
		model.NewPair("B", "C"),
		model.NewPair("N", "O"),
	}

	pas := []*model.PriceAggregate{
		newTestPricePointAggregate(1, "e-a", "B", "C", 2, 1),
		newTestPricePointAggregate(2, "e-b", "B", "C", 4, 1),
		newTestPricePointAggregate(3, "e-c", "A", "C", 8, 1),
		newTestPricePointAggregate(4, "e-d", "A", "B", 16, 1),
		newTestPricePointAggregate(5, "e-e", "D", "B", 32, 1),
		newTestPricePointAggregate(5, "e-f", "E", "B", 64, 1),
		newTestPricePointAggregate(5, "e-z", "E", "B", 128, 1),
		newTestPricePointAggregate(6, "e-x", "X", "Y", 1, 1),
	}

	gofer.processor = &mock.Processor{
		Returns: pas,
		Pairs:   append(pairs, nil),
	}

	ppps := gofer.aggregator.GetSources([]*model.Pair{model.NewPair("B", "C")})

	assert.ElementsMatch(t, []*model.PotentialPricePoint{
		{ Exchange: &model.Exchange{ Name: "e-a" }, Pair: model.NewPair("b", "c") },
		{ Exchange: &model.Exchange{ Name: "e-b" }, Pair: model.NewPair("b", "c") },
	}, ppps)

	for i := 0; i < 1; i++ {
		res, err := gofer.Prices(
			model.NewPair("A", "C"),
			model.NewPair("B", "C"),
			model.NewPair("D", "E"),
			model.NewPair("X", "Y"),
		)
		assert.NoError(t, err)
		assert.NotNil(t, res)

		resBC := res[*model.NewPair("B", "C")]
		assert.NotNil(t, resBC)
		assert.Equal(t, 2.0 + (4 - 2) / 2, resBC.Price)

		resAC := res[*model.NewPair("A", "C")]
		assert.NotNil(t, resAC)
		assert.Equal(t, 8 + (16 * 3.0 - 8) / 2, resAC.Price)

		resDE := res[*model.NewPair("D", "E")]
		assert.NotNil(t, resDE)
		assert.Equal(t, 64.0 / 32, resDE.Price)

		resXY := res[*model.NewPair("X", "Y")]
		assert.Nil(t, resXY, "no explicit price model for X/Y even though source price exists")
	}
}
