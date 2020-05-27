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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/makerdao/gofer/aggregator"
	. "github.com/makerdao/gofer/model"
	"testing"
)

type mockAggregator struct {
	returns map[Pair]*PriceAggregate
}

func (mr *mockAggregator) Ingest(pa *PriceAggregate) {
}

func (mr *mockAggregator) Aggregate(pair *Pair) *PriceAggregate {
	if pair == nil {
		return nil
	}
	return mr.returns[*pair]
}

type mockPather struct {
	ppaths map[Pair][]*PricePath
	pairs  []*Pair
}

func (mp *mockPather) Pairs() []*Pair {
	return mp.pairs
}

func (mp *mockPather) Path(pair *Pair) []*PricePath {
	return mp.ppaths[*pair]
}

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type GoferLibSuite struct {
	suite.Suite
	config     *Config
	pather     *mockPather
	aggregator *mockAggregator
	sources    []*PotentialPricePoint
	processor  *mockProcessor
}

func (suite *GoferLibSuite) TestGoferLibPrices() {
	t := suite.T()

	lib := NewGofer(suite.config)

	pair := Pair{Base: "a", Quote: "b"}
	prices, err := lib.Prices(&pair)
	assert.NoError(t, err)
	assert.Nil(t, prices[pair])

	pair = Pair{Base: "a", Quote: "d"}
	prices, err = lib.Prices(&pair)
	assert.NoError(t, err)
	assert.Equal(t, 0.123, prices[pair].Price)

	suite.processor.returnsErr = fmt.Errorf("processor error")
	_, err = lib.Prices(&pair)
	assert.Error(t, err)
}

func (suite *GoferLibSuite) TestGoferLibPaths() {
	t := suite.T()

	lib := NewGofer(suite.config)

	ppaths := lib.Paths(&Pair{Base: "a", Quote: "d"}, &Pair{Base: "x", Quote: "y"})
	assert.Len(t, ppaths, 1)
	assert.Nil(t, ppaths[Pair{Base: "x", Quote: "y"}])
	assert.Nil(t, ppaths[Pair{Base: "x", Quote: "z"}])
	assert.NotNil(t, ppaths[Pair{Base: "a", Quote: "d"}])
	assert.Equal(t, suite.pather.ppaths[Pair{Base: "a", Quote: "d"}], ppaths[Pair{Base: "a", Quote: "d"}])
}

func (suite *GoferLibSuite) TestGoferLibExchanges() {
	t := suite.T()

	lib := NewGofer(suite.config)

	exchanges := lib.Exchanges(&Pair{Base: "a", Quote: "d"}, &Pair{Base: "x", Quote: "y"})
	assert.ElementsMatch(t, []*Exchange{&Exchange{Name: "exchange-a"}, &Exchange{Name: "exchange-c"}}, exchanges)

	exchanges = lib.Exchanges()
	assert.Len(t, exchanges, 3)

	exchanges = lib.Exchanges(&Pair{Base: "x", Quote: "y"})
	assert.Len(t, exchanges, 0)
}

func (suite *GoferLibSuite) TestGoferLibPairs() {
	t := suite.T()

	lib := NewGofer(suite.config)

	pairs := lib.Pairs()
	assert.ElementsMatch(t, []*Pair{&Pair{Base: "a", Quote: "d"}}, pairs)
}

// Runs before each test
func (suite *GoferLibSuite) SetupTest() {
	sources := []*PotentialPricePoint{
		newPotentialPricePoint("exchange-a", NewPair("a", "b")),
		newPotentialPricePoint("exchange-a", NewPair("a", "z")),
		newPotentialPricePoint("exchange-b", NewPair("a", "z")),
		newPotentialPricePoint("exchange-c", NewPair("b", "d")),
	}
	agg := &mockAggregator{
		returns: map[Pair]*PriceAggregate{
			{Base: "a", Quote: "d"}: {
				PricePoint: &PricePoint{
					Price: 0.123,
				},
			},
			{Base: "e", Quote: "f"}: {
				PricePoint: &PricePoint{
					Price: 0xef,
				},
			},
		},
	}

	newAggregator := func(ppathss []*PricePath) aggregator.Aggregator {
		return agg
	}

	pather := &mockPather{
		ppaths: map[Pair][]*PricePath{
			{Base: "e", Quote: "f"}: []*PricePath{
				&PricePath{
					NewPair("e", "x"),
					NewPair("x", "f"),
				},
			},
			{Base: "a", Quote: "d"}: []*PricePath{
				&PricePath{
					NewPair("a", "b"),
					NewPair("b", "d"),
				},
				&PricePath{
					NewPair("a", "c"),
					NewPair("c", "d"),
				},
			},
		},
		pairs: []*Pair{
			NewPair("a", "d"),
			NewPair("e", "f"),
		},
	}

	processor := &mockProcessor{}

	suite.config = NewConfig(sources, newAggregator, pather, processor)
	suite.sources = sources
	suite.aggregator = agg
	suite.pather = pather
	suite.processor = processor
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestGoferLibSuite(t *testing.T) {
	suite.Run(t, &GoferLibSuite{})
}
