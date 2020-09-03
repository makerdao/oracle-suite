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

	"github.com/makerdao/gofer/internal/query"
	"github.com/makerdao/gofer/pkg/aggregator"
	"github.com/makerdao/gofer/pkg/exchange"
	"github.com/makerdao/gofer/pkg/model"

	"github.com/stretchr/testify/suite"
)

func newPotentialPricePoint(exchangeName string, pair *model.Pair) *model.PotentialPricePoint {
	return &model.PotentialPricePoint{
		Exchange: &model.Exchange{
			Name: exchangeName,
		},
		Pair: pair,
	}
}

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type ProcessorSuite struct {
	suite.Suite
}

// All methods that begin with "Test" are run as tests within a
// suite.
func (suite *ProcessorSuite) TestNegativeProcessOne() {
	set := exchange.NewSet(map[string]exchange.Handler{})

	pair := &model.Pair{
		Base:  "BTC",
		Quote: "ETH",
	}

	p := NewProcessor(set)
	resp, err := p.processOne(&model.PotentialPricePoint{})
	suite.Nil(resp)
	suite.Error(err)

	wrongPp := newPotentialPricePoint("nonexisting", pair)
	p = NewProcessor(set)
	resp, err = p.processOne(wrongPp)
	suite.Nil(resp)
	suite.Error(err)
}

func (suite *ProcessorSuite) TestProcessorProcessOneSuccess() {
	wp := query.NewMockWorkerPool()
	set := exchange.NewSet(map[string]exchange.Handler{
		"binance": &exchange.Binance{Pool: wp},
	})

	pair := &model.Pair{
		Base:  "BTC",
		Quote: "ETH",
	}
	pp := newPotentialPricePoint("binance", pair)
	resp := &query.HTTPResponse{
		Body: []byte(`{"price":"1"}`),
	}
	wp.MockResp(resp)
	p := NewProcessor(set)
	point, err := p.processOne(pp)

	suite.NoError(err)
	suite.EqualValues(pp.Pair, point.Pair)
	suite.EqualValues(1.0, point.Price)
}

func (suite *ProcessorSuite) TestProcessorProcessSuccess() {
	wp := query.NewMockWorkerPool()
	set := exchange.NewSet(map[string]exchange.Handler{
		"binance": &exchange.Binance{Pool: wp},
	})

	pair := &model.Pair{
		Base:  "BTC",
		Quote: "ETH",
	}
	pp := newPotentialPricePoint("binance", pair)
	pp2 := newPotentialPricePoint("binance", pair)
	agg := aggregator.NewMedian([]*model.PotentialPricePoint{pp, pp2}, 1000)

	resp := &query.HTTPResponse{
		Body: []byte(`{"price":"1"}`),
	}
	wp.MockResp(resp)
	p := NewProcessor(set)
	aggr, err := p.Process([]*model.Pair{pair, pair}, agg)

	suite.NoError(err)
	suite.Equal(agg, aggr)

	point := agg.Aggregate(pair)
	suite.NotNil(point)

	suite.EqualValues(pp.Pair, point.Pair)
	suite.EqualValues(1.0, point.Price)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestProcessorSuite(t *testing.T) {
	suite.Run(t, new(ProcessorSuite))
}
