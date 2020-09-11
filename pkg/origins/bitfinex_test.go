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

package origins

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/makerdao/gofer/internal/query"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type BitfinexSuite struct {
	suite.Suite
	pool     query.WorkerPool
	exchange *Bitfinex
}

func (suite *BitfinexSuite) Exchange() Handler {
	return suite.exchange
}

// Setup exchange
func (suite *BitfinexSuite) SetupSuite() {
	suite.exchange = &Bitfinex{Pool: query.NewMockWorkerPool()}
}

func (suite *BitfinexSuite) TearDownTest() {
	// cleanup created pool from prev test
	if suite.pool != nil {
		suite.pool = nil
	}
}

func (suite *BitfinexSuite) TestLocalPair() {
	suite.EqualValues("BTCETH", suite.exchange.localPairName(Pair{Base: "BTC", Quote: "ETH"}))
	suite.EqualValues("BTCUSD", suite.exchange.localPairName(Pair{Base: "BTC", Quote: "USD"}))
	suite.EqualValues("BTCUSD", suite.exchange.localPairName(Pair{Base: "BTC", Quote: "USDT"}))
}

func (suite *BitfinexSuite) TestFailOnWrongInput() {
	// wrong pair
	cr := suite.exchange.Fetch([]Pair{{}})
	suite.Error(cr[0].Error)

	pair := Pair{Base: "BTC", Quote: "ETH"}
	// nil as response
	cr = suite.exchange.Fetch([]Pair{pair})
	suite.Equal(errEmptyExchangeResponse, cr[0].Error)

	// error in response
	ourErr := fmt.Errorf("error")
	resp := &query.HTTPResponse{
		Error: ourErr,
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr = suite.exchange.Fetch([]Pair{pair})
	suite.Equal(ourErr, cr[0].Error)

	// Error unmarshal
	resp = &query.HTTPResponse{
		Body: []byte(""),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr = suite.exchange.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`[0,0]`),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr = suite.exchange.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)
}

func (suite *BitfinexSuite) TestSuccessResponse() {
	pair := Pair{Base: "BTC", Quote: "ETH"}
	resp := &query.HTTPResponse{
		Body: []byte(`[1,1,1,1,1,1,1,1]`),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr := suite.exchange.Fetch([]Pair{pair})
	suite.NoError(cr[0].Error)
	suite.Equal(1.0, cr[0].Tick.Price)
	suite.Greater(cr[0].Tick.Timestamp.Unix(), int64(0))
}

func (suite *BitfinexSuite) TestRealAPICall() {
	testRealAPICall(suite, &Bitfinex{Pool: query.NewHTTPWorkerPool(1)}, "ETH", "BTC")
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestBitfinexSuite(t *testing.T) {
	suite.Run(t, new(BitfinexSuite))
}
