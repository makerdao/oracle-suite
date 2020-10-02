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

	"github.com/makerdao/gofer/internal/query"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type CoinbaseSuite struct {
	suite.Suite
	pool   query.WorkerPool
	origin *Coinbase
}

func (suite *CoinbaseSuite) Origin() Handler {
	return suite.origin
}

// Setup origin
func (suite *CoinbaseSuite) SetupSuite() {
	suite.origin = &Coinbase{Pool: query.NewMockWorkerPool()}
}

func (suite *CoinbaseSuite) TearDownTest() {
	// cleanup created pool from prev test
	if suite.pool != nil {
		suite.pool = nil
	}
}

func (suite *CoinbaseSuite) TestLocalPair() {
	suite.EqualValues("BTC-ETH", suite.origin.localPairName(Pair{Base: "BTC", Quote: "ETH"}))
	suite.EqualValues("BTC-USDC", suite.origin.localPairName(Pair{Base: "BTC", Quote: "USD"}))
}

func (suite *CoinbaseSuite) TestFailOnWrongInput() {
	// wrong pair
	cr := suite.origin.Fetch([]Pair{{}})
	suite.Error(cr[0].Error)

	pair := Pair{Base: "BTC", Quote: "ETH"}
	// nil as response
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Equal(errEmptyOriginResponse, cr[0].Error)

	// error in response
	ourErr := fmt.Errorf("error")
	resp := &query.HTTPResponse{
		Error: ourErr,
	}
	suite.origin.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Equal(ourErr, cr[0].Error)

	// Error unmarshal
	resp = &query.HTTPResponse{
		Body: []byte(""),
	}
	suite.origin.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`{"price":"abc"}`),
	}
	suite.origin.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`{"price":"1","ask":"abc"}`),
	}
	suite.origin.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`{"price":"1","ask":"1","volume":"abc"}`),
	}
	suite.origin.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`{"price":"1","ask":"1","volume":"1","bid":"abc"}`),
	}
	suite.origin.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)
}

func (suite *CoinbaseSuite) TestSuccessResponse() {
	pair := Pair{Base: "BTC", Quote: "ETH"}
	resp := &query.HTTPResponse{
		Body: []byte(`{"price":"1","ask":"2","volume":"3","bid":"4"}`),
	}
	suite.origin.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr := suite.origin.Fetch([]Pair{pair})
	suite.NoError(cr[0].Error)
	suite.Equal(1.0, cr[0].Tick.Price)
	suite.Equal(2.0, cr[0].Tick.Ask)
	suite.Equal(3.0, cr[0].Tick.Volume24h)
	suite.Equal(4.0, cr[0].Tick.Bid)
	suite.Greater(cr[0].Tick.Timestamp.Unix(), int64(2))
}

func (suite *CoinbaseSuite) TestRealAPICall() {
	testRealAPICall(suite, &Coinbase{Pool: query.NewHTTPWorkerPool(1)}, "ETH", "BTC")
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestCoinbaseSuite(t *testing.T) {
	suite.Run(t, new(CoinbaseSuite))
}
