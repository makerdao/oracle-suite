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

package exchange

import (
	"fmt"
	"testing"

	"github.com/makerdao/gofer/model"
	"github.com/makerdao/gofer/query"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type UniswapSuite struct {
	suite.Suite
	pool     query.WorkerPool
	exchange Handler
}

func (suite *UniswapSuite) Exchange() Handler {
	return suite.exchange
}

// Setup exchange
func (suite *UniswapSuite) SetupSuite() {
	suite.exchange = &Uniswap{}
}

func (suite *UniswapSuite) TearDownTest() {
	// cleanup created pool from prev test
	if suite.pool != nil {
		suite.pool = nil
	}
}

func (suite *UniswapSuite) TestLocalPair() {
	suite.EqualValues("0xcffdded873554f362ac02f8fb1f02e5ada10516f", suite.exchange.LocalPairName(model.NewPair("COMP", "ETH")))
	suite.EqualValues("0x8878df9e1a7c87dcbf6d3999d997f262c05d8c70", suite.exchange.LocalPairName(model.NewPair("LRC", "ETH")))
	suite.EqualValues("0xf49c43ae0faf37217bdcb00df478cf793edd6687", suite.exchange.LocalPairName(model.NewPair("KNC", "ETH")))
}

func (suite *UniswapSuite) TestFailOnWrongInput() {
	// no pool
	_, err := suite.exchange.Call(nil, nil)
	suite.Equal(errNoPoolPassed, err)

	// empty pp
	_, err = suite.exchange.Call(newMockWorkerPool(nil), nil)
	suite.Error(err)

	// wrong pp
	_, err = suite.exchange.Call(newMockWorkerPool(nil), &model.PotentialPricePoint{})
	suite.Error(err)

	pp := newPotentialPricePoint("uniswap", "COMP", "ETH")
	// nil as response
	_, err = suite.exchange.Call(newMockWorkerPool(nil), pp)
	suite.Equal(errEmptyExchangeResponse, err)

	// error in response
	ourErr := fmt.Errorf("error")
	resp := &query.HTTPResponse{
		Error: ourErr,
	}
	_, err = suite.exchange.Call(newMockWorkerPool(resp), pp)
	suite.Equal(ourErr, err)

	// Error unmarshal
	resp = &query.HTTPResponse{
		Body: []byte(""),
	}
	_, err = suite.exchange.Call(newMockWorkerPool(resp), pp)
	suite.Error(err)

	// Error unmarshal
	resp = &query.HTTPResponse{
		Body: []byte("{}"),
	}
	_, err = suite.exchange.Call(newMockWorkerPool(resp), pp)
	suite.Error(err)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`{"data":{}`),
	}
	_, err = suite.exchange.Call(newMockWorkerPool(resp), pp)
	suite.Error(err)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`{"data":{"pairs":[]}}`),
	}
	_, err = suite.exchange.Call(newMockWorkerPool(resp), pp)
	suite.Error(err)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`{"data":{"pairs":[{}]}}`),
	}
	_, err = suite.exchange.Call(newMockWorkerPool(resp), pp)
	suite.Error(err)
}

func (suite *UniswapSuite) TestSuccessResponse() {
	pp := newPotentialPricePoint("uniswap", "COMP", "ETH")
	resp := &query.HTTPResponse{
		Body: []byte(`{"data":{"pairs":[{"token0Price":"0", "token1Price":"1"}]}}`),
	}
	point, err := suite.exchange.Call(newMockWorkerPool(resp), pp)
	suite.NoError(err)
	suite.Equal(pp.Exchange, point.Exchange)
	suite.Equal(pp.Pair, point.Pair)
	suite.Equal(1.0, point.Price)
}

func (suite *UniswapSuite) TestSuccessResponseForToken0Price() {
	pp := newPotentialPricePoint("uniswap", "KNC", "ETH")
	resp := &query.HTTPResponse{
		Body: []byte(`{"data":{"pairs":[{"token0Price":"1", "token1Price":"2"}]}}`),
	}
	point, err := suite.exchange.Call(newMockWorkerPool(resp), pp)
	suite.NoError(err)
	suite.Equal(pp.Exchange, point.Exchange)
	suite.Equal(pp.Pair, point.Pair)
	suite.Equal(1.0, point.Price)
}

func (suite *UniswapSuite) TestRealAPICall() {
	testRealAPICall(suite, "COMP", "ETH")
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestUniswapSuiteSuite(t *testing.T) {
	suite.Run(t, new(UniswapSuite))
}
