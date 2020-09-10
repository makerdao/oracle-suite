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

	"github.com/makerdao/gofer/internal/query"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type UniswapSuite struct {
	suite.Suite
	pool     query.WorkerPool
	exchange *Uniswap
}

func (suite *UniswapSuite) Exchange() Handler {
	return suite.exchange
}

// Setup exchange
func (suite *UniswapSuite) SetupSuite() {
	suite.exchange = &Uniswap{Pool: query.NewMockWorkerPool()}
}

func (suite *UniswapSuite) TearDownTest() {
	// cleanup created pool from prev test
	if suite.pool != nil {
		suite.pool = nil
	}
}

func (suite *UniswapSuite) TestLocalPair() {
	suite.EqualValues("0xcffdded873554f362ac02f8fb1f02e5ada10516f", suite.exchange.localPairName(Pair{Base: "COMP", Quote: "ETH"}))
	suite.EqualValues("0x8878df9e1a7c87dcbf6d3999d997f262c05d8c70", suite.exchange.localPairName(Pair{Base: "LRC", Quote: "ETH"}))
	suite.EqualValues("0xf49c43ae0faf37217bdcb00df478cf793edd6687", suite.exchange.localPairName(Pair{Base: "KNC", Quote: "ETH"}))
}

func (suite *UniswapSuite) TestFailOnWrongInput() {
	// wrong pp
	cr := suite.exchange.Call([]Pair{{}})
	suite.Error(cr[0].Error)

	pp := Pair{Base: "COMP", Quote: "ETH"}
	// nil as response
	cr = suite.exchange.Call([]Pair{pp})
	suite.Equal(errEmptyExchangeResponse, cr[0].Error)

	// error in response
	ourErr := fmt.Errorf("error")
	resp := &query.HTTPResponse{
		Error: ourErr,
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr = suite.exchange.Call([]Pair{pp})
	suite.Equal(ourErr, cr[0].Error)

	// Error unmarshal
	resp = &query.HTTPResponse{
		Body: []byte(""),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr = suite.exchange.Call([]Pair{pp})
	suite.Error(cr[0].Error)

	// Error unmarshal
	resp = &query.HTTPResponse{
		Body: []byte("{}"),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr = suite.exchange.Call([]Pair{pp})
	suite.Error(cr[0].Error)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`{"data":{}`),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr = suite.exchange.Call([]Pair{pp})
	suite.Error(cr[0].Error)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`{"data":{"pairs":[]}}`),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr = suite.exchange.Call([]Pair{pp})
	suite.Error(cr[0].Error)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`{"data":{"pairs":[{}]}}`),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr = suite.exchange.Call([]Pair{pp})
	suite.Error(cr[0].Error)
}

func (suite *UniswapSuite) TestSuccessResponse() {
	pp := Pair{Base: "COMP", Quote: "ETH"}
	resp := &query.HTTPResponse{
		Body: []byte(`{"data":{"pairs":[{"token0Price":"0", "token1Price":"1"}]}}`),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr := suite.exchange.Call([]Pair{pp})
	suite.NoError(cr[0].Error)
	suite.Equal(1.0, cr[0].Tick.Price)
}

func (suite *UniswapSuite) TestSuccessResponseForToken0Price() {
	pp := Pair{Base: "KNC", Quote: "ETH"}
	resp := &query.HTTPResponse{
		Body: []byte(`{"data":{"pairs":[{"token0Price":"1", "token1Price":"2"}]}}`),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr := suite.exchange.Call([]Pair{pp})
	suite.NoError(cr[0].Error)
	suite.Equal(1.0, cr[0].Tick.Price)
}

func (suite *UniswapSuite) TestRealAPICall() {
	testRealAPICall(suite, &Uniswap{Pool: query.NewHTTPWorkerPool(1)}, "COMP", "ETH")
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestUniswapSuiteSuite(t *testing.T) {
	suite.Run(t, new(UniswapSuite))
}
