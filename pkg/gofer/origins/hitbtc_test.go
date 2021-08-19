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

	"github.com/makerdao/oracle-suite/internal/query"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type HitbtcSuite struct {
	suite.Suite
	pool   query.WorkerPool
	origin *BaseExchangeHandler
}

func (suite *HitbtcSuite) Origin() Handler {
	return suite.origin
}

// Setup exchange
func (suite *HitbtcSuite) SetupSuite() {
	suite.origin = NewBaseExchangeHandler(Hitbtc{WorkerPool: query.NewMockWorkerPool()}, nil)
}

func (suite *HitbtcSuite) TearDownTest() {
	// cleanup created pool from prev test
	if suite.pool != nil {
		suite.pool = nil
	}
}

func (suite *HitbtcSuite) TestLocalPair() {
	ex := suite.origin.ExchangeHandler.(Hitbtc)
	suite.EqualValues("BTCETH", ex.localPairName(Pair{Base: "BTC", Quote: "ETH"}))
	suite.EqualValues("BTCUSD", ex.localPairName(Pair{Base: "BTC", Quote: "USD"}))
}

func (suite *HitbtcSuite) TestFailOnWrongInput() {
	pair := Pair{Base: "BTC", Quote: "ETH"}
	// nil as response
	cr := suite.origin.Fetch([]Pair{pair})
	suite.Equal(ErrEmptyOriginResponse, cr[0].Error)

	// error in response
	ourErr := fmt.Errorf("error")
	resp := &query.HTTPResponse{
		Error: ourErr,
	}
	suite.origin.Pool().(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Equal(ourErr, cr[0].Error)

	// Error unmarshal
	resp = &query.HTTPResponse{
		Body: []byte(""),
	}
	suite.origin.Pool().(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`[{"last":"abc"}]`),
	}
	suite.origin.Pool().(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`[{"last":"1","ask":"abc"}]`),
	}
	suite.origin.Pool().(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`[{"last":"1","ask":"1","volume":"abc"}]`),
	}
	suite.origin.Pool().(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`[{"last":"1","ask":"1","volume":"1","bid":"abc"}]`),
	}
	suite.origin.Pool().(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`[{"last":"1","ask":"1","volume":"1","bid":"abc","symbol":"abc"}]`),
	}
	suite.origin.Pool().(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`{"last":"1","ask":"2","volume":"3","bid":"4","symbol":"BTCETH","timestamp":"2020-04-24T20:09:36.229Z"}`),
	}
	suite.origin.Pool().(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)
}

func (suite *HitbtcSuite) TestSuccessResponse() {
	// Empty fetch.
	cr := suite.origin.Fetch([]Pair{})
	suite.Len(cr, 0)

	pair := Pair{Base: "BTC", Quote: "ETH"}
	resp := &query.HTTPResponse{
		Body: []byte(`[{"last":"1","ask":"2","volume":"3","bid":"4","symbol":"BTCETH","timestamp":"2020-04-24T20:09:36.229Z"}]`),
	}
	suite.origin.Pool().(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.NoError(cr[0].Error)
	suite.Equal(1.0, cr[0].Price.Price)
	suite.Equal(2.0, cr[0].Price.Ask)
	suite.Equal(3.0, cr[0].Price.Volume24h)
	suite.Equal(4.0, cr[0].Price.Bid)
	suite.Equal(cr[0].Price.Timestamp.Unix(), int64(1587758976))
}

func (suite *HitbtcSuite) TestRealAPICall() {
	//hitbtc := &Hitbtc{Pool: query.NewHTTPWorkerPool(1)}
	//testRealAPICall(suite, hitbtc, "ETH", "BTC")
	//testRealBatchAPICall(suite, hitbtc, []Pair{
	//	{Base: "BTC", Quote: "USD"},
	//	{Base: "DOGE", Quote: "BTC"},
	//	{Base: "REP", Quote: "USDT"},
	//})
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestHitbtcSuite(t *testing.T) {
	suite.Run(t, new(HitbtcSuite))
}
