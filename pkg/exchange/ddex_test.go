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
type DdexSuite struct {
	suite.Suite
	pool     query.WorkerPool
	exchange *Ddex
}

func (suite *DdexSuite) Exchange() Handler {
	return suite.exchange
}

// Setup exchange
func (suite *DdexSuite) SetupSuite() {
	suite.exchange = &Ddex{Pool: query.NewMockWorkerPool()}
}

func (suite *DdexSuite) TearDownTest() {
	// cleanup created pool from prev test
	if suite.pool != nil {
		suite.pool = nil
	}
}

func (suite *DdexSuite) TestLocalPair() {
	suite.EqualValues("BTC-ETH", suite.exchange.localPairName(Pair{Base: "BTC", Quote: "ETH"}))
	suite.NotEqual("BTC-USDC", suite.exchange.localPairName(Pair{Base: "BTC", Quote: "USD"}))
}

func (suite *DdexSuite) TestFailOnWrongInput() {
	// wrong pair
	cr := suite.exchange.Call([]Pair{{}})
	suite.Error(cr[0].Error)

	pair := Pair{Base: "BTC", Quote: "ETH"}
	// nil as response
	cr = suite.exchange.Call([]Pair{pair})
	suite.Equal(errEmptyExchangeResponse, cr[0].Error)

	// error in response
	ourErr := fmt.Errorf("error")
	resp := &query.HTTPResponse{
		Error: ourErr,
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr = suite.exchange.Call([]Pair{pair})
	suite.Equal(ourErr, cr[0].Error)

	// Error unmarshal
	resp = &query.HTTPResponse{
		Body: []byte(""),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr = suite.exchange.Call([]Pair{pair})
	suite.Error(cr[0].Error)

	for n, r := range [][]byte{
		// invalid desc
		[]byte(`{
		   "status":0,
		   "desc":"err",
		   "template":"",
		   "params":null,
		   "data":{
			  "orderbook":{
				 "sequence":143661147,
				 "marketId":"WBTC-USDT",
				 "bids":[
					{
					   "price":"11691.6",
					   "amount":"0.4173"
					}
				 ],
				 "asks":[
					{
					   "price":"11719.8",
					   "amount":"0.3709"
					}
				 ]
			  }
		   }
		}`),
		// invalid ask price
		[]byte(`{
		   "status":0,
		   "desc":"success",
		   "template":"",
		   "params":null,
		   "data":{
			  "orderbook":{
				 "sequence":143661147,
				 "marketId":"WBTC-USDT",
				 "bids":[
					{
					   "price":"11691.6",
					   "amount":"0.4173"
					}
				 ],
				 "asks":[
					{
					   "price":"err",
					   "amount":"0.3709"
					}
				 ]
			  }
		   }
		}`),
		// invalid bid price
		[]byte(`{
		   "status":0,
		   "desc":"success",
		   "template":"",
		   "params":null,
		   "data":{
			  "orderbook":{
				 "sequence":143661147,
				 "marketId":"WBTC-USDT",
				 "bids":[
					{
					   "price":"err",
					   "amount":"0.4173"
					}
				 ],
				 "asks":[
					{
					   "price":"11719.8",
					   "amount":"0.3709"
					}
				 ]
			  }
		   }
		}`),
		// empty order book
		[]byte(`{
		   "status":0,
		   "desc":"success",
		   "template":"",
		   "params":null,
		   "data":{
			  "orderbook":{
				 "sequence":143661147,
				 "marketId":"WBTC-USDT",
				 "bids":[],
				 "asks":[]
			  }
		   }
		}`),
	} {
		suite.T().Run(fmt.Sprintf("Case-%d", n+1), func(t *testing.T) {
			resp = &query.HTTPResponse{Body: r}
			suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
			cr = suite.exchange.Call([]Pair{pair})
			suite.Error(cr[0].Error)
		})
	}
}

func (suite *DdexSuite) TestSuccessResponse() {
	pair := Pair{Base: "BTC", Quote: "ETH"}
	resp := &query.HTTPResponse{
		Body: []byte(`{
		   "status":0,
		   "desc":"success",
		   "template":"",
		   "params":null,
		   "data":{
			  "orderbook":{
				 "sequence":143661147,
				 "marketId":"WBTC-USDT",
				 "bids":[
					{
					   "price":"100.5",
					   "amount":"0.4173"
					}
				 ],
				 "asks":[
					{
					   "price":"101.6",
					   "amount":"0.3709"
					}
				 ]
			  }
		   }
		}`),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr := suite.exchange.Call([]Pair{pair})
	suite.NoError(cr[0].Error)
	suite.Equal(101.6, cr[0].Tick.Ask)
	suite.Equal(100.5, cr[0].Tick.Bid)
	suite.Equal(100.5, cr[0].Tick.Price)
}

func (suite *DdexSuite) TestRealAPICall() {
	testRealAPICall(suite, &Ddex{Pool: query.NewHTTPWorkerPool(1)}, "WBTC", "USDT")
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestDdexSuite(t *testing.T) {
	suite.Run(t, new(DdexSuite))
}
