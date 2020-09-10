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
type CoinbaseProSuite struct {
	suite.Suite
	pool     query.WorkerPool
	exchange *CoinbasePro
}

func (suite *CoinbaseProSuite) Exchange() Handler {
	return suite.exchange
}

// Setup exchange
func (suite *CoinbaseProSuite) SetupSuite() {
	suite.exchange = &CoinbasePro{Pool: query.NewMockWorkerPool()}
}

func (suite *CoinbaseProSuite) TearDownTest() {
	// cleanup created pool from prev test
	if suite.pool != nil {
		suite.pool = nil
	}
}

func (suite *CoinbaseProSuite) TestLocalPair() {
	suite.EqualValues("BTC-ETH", suite.exchange.localPairName(Pair{Base: "BTC", Quote: "ETH"}))
	suite.EqualValues("BTC-USD", suite.exchange.localPairName(Pair{Base: "BTC", Quote: "USD"}))
}

func (suite *CoinbaseProSuite) TestFailOnWrongInput() {
	// wrong pp
	cr := suite.exchange.Call([]Pair{{}})
	suite.Error(cr[0].Error)

	pp := Pair{Base: "BTC", Quote: "ETH"}
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

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`{"price":"abc"}`),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr = suite.exchange.Call([]Pair{pp})
	suite.Error(cr[0].Error)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`{"price":"1","ask":"abc"}`),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr = suite.exchange.Call([]Pair{pp})
	suite.Error(cr[0].Error)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`{"price":"1","ask":"1","volume":"abc"}`),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr = suite.exchange.Call([]Pair{pp})
	suite.Error(cr[0].Error)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`{"price":"1","ask":"1","volume":"1","bid":"abc"}`),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr = suite.exchange.Call([]Pair{pp})
	suite.Error(cr[0].Error)
}

func (suite *CoinbaseProSuite) TestSuccessResponse() {
	pp := Pair{Base: "BTC", Quote: "ETH"}
	resp := &query.HTTPResponse{
		Body: []byte(`{"price":"1","ask":"2","volume":"3","bid":"4"}`),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr := suite.exchange.Call([]Pair{pp})
	suite.NoError(cr[0].Error)
	suite.Equal(1.0, cr[0].Tick.Price)
	suite.Equal(2.0, cr[0].Tick.Ask)
	suite.Equal(3.0, cr[0].Tick.Volume24h)
	suite.Equal(4.0, cr[0].Tick.Bid)
	suite.Greater(cr[0].Tick.Timestamp.Unix(), int64(2))
}

func (suite *CoinbaseProSuite) TestRealAPICall() {
	testRealAPICall(suite, &CoinbasePro{Pool: query.NewHTTPWorkerPool(1)}, "ETH", "BTC")
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestCoinbaseProSuite(t *testing.T) {
	suite.Run(t, new(CoinbaseProSuite))
}
