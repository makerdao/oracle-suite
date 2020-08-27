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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	aggregatorMock "github.com/makerdao/gofer/internal/pkg/mock/aggregator"
	processorMock "github.com/makerdao/gofer/internal/pkg/mock/prcessor"
	"github.com/makerdao/gofer/pkg/model"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type GoferLibSuite struct {
	suite.Suite
	aggregator *aggregatorMock.Aggregator
	sources    []*model.PotentialPricePoint
	processor  *processorMock.Processor
}

func (suite *GoferLibSuite) TestGoferLibPrices() {
	t := suite.T()

	lib := NewGofer(suite.aggregator, suite.processor)

	pair := model.NewPair("a", "b")
	prices, err := lib.Prices(pair)
	assert.NoError(t, err)
	assert.Nil(t, prices[*pair])

	pair = model.NewPair("a", "d")
	prices, err = lib.Prices(pair)
	assert.NoError(t, err)
	assert.Equal(t, 0.123, prices[*pair].Price)

	suite.processor.ReturnsErr = fmt.Errorf("processor error")
	_, err = lib.Prices(pair)
	assert.Error(t, err)
}

func (suite *GoferLibSuite) TestGoferLibExchanges() {
	t := suite.T()

	lib := NewGofer(suite.aggregator, suite.processor)

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

// Runs before each test
func (suite *GoferLibSuite) SetupTest() {
	sources := []*model.PotentialPricePoint{
		newPotentialPricePoint("exchange-a", model.NewPair("a", "b")),
		newPotentialPricePoint("exchange-a", model.NewPair("a", "z")),
		newPotentialPricePoint("exchange-b", model.NewPair("a", "z")),
		newPotentialPricePoint("exchange-c", model.NewPair("b", "d")),
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
		Sources: map[model.Pair][]*model.PotentialPricePoint{
			*model.NewPair("a", "b"): []*model.PotentialPricePoint{
				newPotentialPricePoint("exchange-a", model.NewPair("a", "b")),
			},
			*model.NewPair("a", "z"): []*model.PotentialPricePoint{
				newPotentialPricePoint("exchange-a", model.NewPair("a", "z")),
				newPotentialPricePoint("exchange-b", model.NewPair("a", "z")),
			},
			*model.NewPair("b", "d"): []*model.PotentialPricePoint{
				newPotentialPricePoint("exchange-c", model.NewPair("b", "d")),
				newPotentialPricePoint("exchange-d", model.NewPair("b", "c")),
				newPotentialPricePoint("exchange-d", model.NewPair("d", "c")),
			},
		},
	}

	processor := &processorMock.Processor{}

	suite.sources = sources
	suite.aggregator = agg
	suite.processor = processor
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestGoferLibSuite(t *testing.T) {
	suite.Run(t, &GoferLibSuite{})
}
