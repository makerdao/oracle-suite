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
	"github.com/makerdao/gofer/aggregator"
	"github.com/makerdao/gofer/model"
	"github.com/makerdao/gofer/query"
	"testing"

	"github.com/stretchr/testify/suite"
)

// mockWorkerPool mock worker pool implementation for tests
type mockWorkerPool struct {
	resp *query.HTTPResponse
}

func newMockWorkerPool(resp *query.HTTPResponse) *mockWorkerPool {
	return &mockWorkerPool{
		resp: resp,
	}
}

func (mwp *mockWorkerPool) Ready() bool {
	return true
}

func (mwp *mockWorkerPool) Start() {}

func (mwp *mockWorkerPool) Stop() error {
	return nil
}

func (mwp *mockWorkerPool) Query(req *query.HTTPRequest) *query.HTTPResponse {
	return mwp.resp
}

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
	pair := &model.Pair{
		Base:  "BTC",
		Quote: "ETH",
	}
	pp := newPotentialPricePoint("coinbase", pair)
	// Wrong worker pool
	p := NewProcessor(nil)
	resp, err := p.ProcessOne(pp)
	suite.Nil(resp)
	suite.Error(err)

	p = NewProcessor(newMockWorkerPool(nil))
	resp, err = p.ProcessOne(&model.PotentialPricePoint{})
	suite.Nil(resp)
	suite.Error(err)

	wrongPp := newPotentialPricePoint("nonexisting", pair)
	p = NewProcessor(newMockWorkerPool(nil))
	resp, err = p.ProcessOne(wrongPp)
	suite.Nil(resp)
	suite.Error(err)
}

func (suite *ProcessorSuite) TestProcessorProcessOneSuccess() {
	pair := &model.Pair{
		Base:  "BTC",
		Quote: "ETH",
	}
	pp := newPotentialPricePoint("binance", pair)
	resp := &query.HTTPResponse{
		Body: []byte(`{"price":"1"}`),
	}
	wp := newMockWorkerPool(resp)
	p := NewProcessor(wp)
	point, err := p.ProcessOne(pp)

	suite.NoError(err)
	suite.EqualValues(pp.Pair, point.Pair)
	suite.EqualValues(1.0, point.Price)
}

func (suite *ProcessorSuite) TestProcessorProcessSuccess() {
	pair := &model.Pair{
		Base:  "BTC",
		Quote: "ETH",
	}
	pp := newPotentialPricePoint("binance", pair)
	pp2 := newPotentialPricePoint("binance", pair)
	agg := aggregator.NewMedian(1000)

	resp := &query.HTTPResponse{
		Body: []byte(`{"price":"1"}`),
	}
	wp := newMockWorkerPool(resp)
	p := NewProcessor(wp)
	aggr, err := p.Process([]*model.PotentialPricePoint{pp, pp2}, agg)

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
