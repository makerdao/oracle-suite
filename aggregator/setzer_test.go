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

func TestSetzerAggregator(t *testing.T) {
	pas := []*model.PriceAggregate{
		newTestPricePointAggregate(0, "e1", "b", "c", 101, 1),
		newTestPricePointAggregate(0, "e2", "b", "c", 102, 1),
		newTestPricePointAggregate(0, "e3", "b", "c", 103, 1),
		newTestPricePointAggregate(0, "e1", "a", "c", 104, 1),
		newTestPricePointAggregate(0, "e1", "a", "b", 105, 1),
		newTestPricePointAggregate(0, "e2", "b", "a", 106, 1),
		newTestPricePointAggregate(0, "e1", "n", "o", 107, 1),
	}

	pmm := PriceModelMap{
		{*model.NewPair("b", "c")}: PriceModel{
			Method: "median",
			Sources: []*PriceRef{
				{ExchangeName: "e1"},
				{ExchangeName: "e2"},
				{ExchangeName: "e3"},
			},
		},
		{*model.NewPair("a", "c")}: PriceModel{
			Method: "median",
			Sources: []*PriceRef{
				{ExchangeName: "e1", Pair: &Pair{*model.NewPair("a", "c")}},
				{
					ExchangeName: "e1",
					Pair:     &Pair{*model.NewPair("a", "b")},
					Ref:      &PriceRef{Pair: &Pair{*model.NewPair("b", "c")}},
					Op:       MULTIPLY,
				},
				{
					ExchangeName: "e2",
					Pair:     &Pair{*model.NewPair("b", "a")},
					Ref:      &PriceRef{Pair: &Pair{*model.NewPair("b", "c")}},
					Op:       DIVIDE,
				},
			},
		},
		{*model.NewPair("x", "y")}: PriceModel{
			Method: "median",
			Sources: []*PriceRef{
				{ExchangeName: "e4", Pair: &Pair{*model.NewPair("x", "y")}},
			},
		},
	}

	exchanges := []*model.Exchange{
		{ Name: "e1" },
		{ Name: "e2" },
		{ Name: "e3", Config: map[string]string{"a": "1"} },
	}

	setz := NewSetz(exchanges, pmm)

	res := setz.Aggregate(nil)
	assert.Nil(t, res)

	res = setz.Aggregate(model.NewPair("x", "y"))
	assert.Nil(t, res)

	res = randomReduce(setz, model.NewPair("a", "c"), pas)
	assert.NotNil(t, res)

	assert.Equal(t, model.NewPair("a", "c"), res.Pair)
	assert.Equal(t, 104.0, res.Price)

	res = randomReduce(setz, model.NewPair("b", "c"), pas)
	assert.NotNil(t, res)

	assert.Equal(t, model.NewPair("b", "c"), res.Pair)
	assert.Equal(t, 102.0, res.Price)

	ppps := setz.GetSources([]*model.Pair{ model.NewPair("b", "c") })
	assert.ElementsMatch(t, []*model.PotentialPricePoint{
		{ Exchange: &model.Exchange{ Name: "e1" }, Pair: model.NewPair("b", "c") },
		{ Exchange: &model.Exchange{ Name: "e2" }, Pair: model.NewPair("b", "c") },
		{ Exchange: &model.Exchange{ Name: "e3", Config: map[string]string{"a": "1"} }, Pair: model.NewPair("b", "c") },
	}, ppps)

	ppps = setz.GetSources([]*model.Pair{ model.NewPair("a", "c") })
	assert.ElementsMatch(t, []*model.PotentialPricePoint{
		{ Exchange: &model.Exchange{ Name: "e1" }, Pair: model.NewPair("b", "c") },
		{ Exchange: &model.Exchange{ Name: "e2" }, Pair: model.NewPair("b", "c") },
		{ Exchange: &model.Exchange{ Name: "e3", Config: map[string]string{"a": "1"} }, Pair: model.NewPair("b", "c") },
		{ Exchange: &model.Exchange{ Name: "e1" }, Pair: model.NewPair("a", "c") },
		{ Exchange: &model.Exchange{ Name: "e1" }, Pair: model.NewPair("a", "b") },
		{ Exchange: &model.Exchange{ Name: "e2" }, Pair: model.NewPair("b", "a") },
	}, ppps)

	ppps = setz.GetSources(nil)
	assert.ElementsMatch(t, []*model.PotentialPricePoint{
		{ Exchange: &model.Exchange{ Name: "e1" }, Pair: model.NewPair("b", "c") },
		{ Exchange: &model.Exchange{ Name: "e2" }, Pair: model.NewPair("b", "c") },
		{ Exchange: &model.Exchange{ Name: "e3", Config: map[string]string{"a": "1"} }, Pair: model.NewPair("b", "c") },
		{ Exchange: &model.Exchange{ Name: "e1" }, Pair: model.NewPair("a", "c") },
		{ Exchange: &model.Exchange{ Name: "e1" }, Pair: model.NewPair("a", "b") },
		{ Exchange: &model.Exchange{ Name: "e2" }, Pair: model.NewPair("b", "a") },
		{ Exchange: &model.Exchange{ Name: "e4" }, Pair: model.NewPair("x", "y") },
	}, ppps)
}
