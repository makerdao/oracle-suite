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
	"github.com/stretchr/testify/suite"

	aggregatorMock "github.com/makerdao/gofer/internal/mock/aggregator"
	"github.com/makerdao/gofer/pkg/model"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type GoferLibSuite struct {
	suite.Suite
	aggregator *aggregatorMock.Aggregator
	sources    []*model.PricePoint
}

func (suite *GoferLibSuite) TestGoferLibExchanges() {
	t := suite.T()

	lib := NewGofer(suite.aggregator, nil)

	exchanges := lib.Exchanges(model.NewPair("a", "b"), model.NewPair("x", "y"))
	assert.Len(t, exchanges, 1)

	exchanges = lib.Exchanges(model.NewPair("a", "b"), model.NewPair("a", "z"))
	assert.ElementsMatch(
		t,
		[]*model.Exchange{
			{Name: "exchange-a"},
			{Name: "exchange-b"},
		},
		exchanges,
	)

	exchanges = lib.Exchanges(model.NewPair("a", "z"))
	assert.ElementsMatch(
		t,
		[]*model.Exchange{
			{Name: "exchange-a"},
			{Name: "exchange-b"},
		},
		exchanges,
	)

	exchanges = lib.Exchanges(model.NewPair("b", "d"))
	assert.ElementsMatch(
		t,
		[]*model.Exchange{
			{Name: "exchange-c"},
			{Name: "exchange-d"},
		},
		exchanges,
	)

	exchanges = lib.Exchanges()
	assert.Len(t, exchanges, 4)

	exchanges = lib.Exchanges(model.NewPair("x", "y"))
	assert.Len(t, exchanges, 0)
}

func newPricePoint(exchangeName string, pair *model.Pair) *model.PricePoint {
	return &model.PricePoint{
		Exchange: &model.Exchange{
			Name: exchangeName,
		},
		Pair: pair,
	}
}

// Runs before each test
func (suite *GoferLibSuite) SetupTest() {
	sources := []*model.PricePoint{
		newPricePoint("exchange-a", model.NewPair("a", "b")),
		newPricePoint("exchange-a", model.NewPair("a", "z")),
		newPricePoint("exchange-b", model.NewPair("a", "z")),
		newPricePoint("exchange-c", model.NewPair("b", "d")),
	}
	agg := &aggregatorMock.Aggregator{
		Returns: map[model.Pair]*model.PriceAggregate{
			*model.NewPair("a", "d"): {
				PricePoint: &model.PricePoint{
					Price: 0.123,
				},
			},
			*model.NewPair("e", "f"): {
				PricePoint: &model.PricePoint{
					Price: 0xef,
				},
			},
		},
		Sources: map[model.Pair][]*model.PricePoint{
			*model.NewPair("a", "b"): {
				newPricePoint("exchange-a", model.NewPair("a", "b")),
			},
			*model.NewPair("a", "z"): {
				newPricePoint("exchange-a", model.NewPair("a", "z")),
				newPricePoint("exchange-b", model.NewPair("a", "z")),
			},
			*model.NewPair("b", "d"): {
				newPricePoint("exchange-c", model.NewPair("b", "d")),
				newPricePoint("exchange-d", model.NewPair("b", "c")),
				newPricePoint("exchange-d", model.NewPair("d", "c")),
			},
		},
	}

	suite.sources = sources
	suite.aggregator = agg
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestGoferLibSuite(t *testing.T) {
	suite.Run(t, &GoferLibSuite{})
}
