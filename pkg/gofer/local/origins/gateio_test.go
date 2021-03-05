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
type GateioSuite struct {
	suite.Suite
	pool   query.WorkerPool
	origin *Gateio
}

func (suite *GateioSuite) Origin() Handler {
	return suite.origin
}

// Setup exchange
func (suite *GateioSuite) SetupSuite() {
	suite.origin = &Gateio{Pool: query.NewMockWorkerPool()}
}

func (suite *GateioSuite) TearDownTest() {
	// cleanup created pool from prev test
	if suite.pool != nil {
		suite.pool = nil
	}
}

func (suite *GateioSuite) TestLocalPair() {
	suite.EqualValues("BTC_ETH", suite.origin.localPairName(Pair{Base: "BTC", Quote: "ETH"}))
	suite.EqualValues("BTC_USD", suite.origin.localPairName(Pair{Base: "BTC", Quote: "USD"}))
}

func (suite *GateioSuite) TestFailOnWrongInput() {
	// No pairs.
	cr := suite.origin.Fetch([]Pair{})
	suite.Len(cr, 0)

	pair := Pair{Base: "BTC", Quote: "ETH"}
	// nil as response
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Equal(ErrEmptyOriginResponse, cr[0].Error)

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

	// Error unmarshal
	resp = &query.HTTPResponse{
		Body: []byte("[{}]"),
	}
	suite.origin.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`[{"last":"abc"}]`),
	}
	suite.origin.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`[{"last":"1","currency_pair":"abc"}]`),
	}
	suite.origin.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)
}

func (suite *GateioSuite) TestSuccessResponse() {
	pair := Pair{Base: "C", Quote: "D"}
	resp := &query.HTTPResponse{
		Body: []byte(`[{"currency_pair":"A_B","last":"1","lowest_ask":"2","highest_bid":"3","quote_volume":"4"},{"currency_pair":"C_D","last":"5","lowest_ask":"6","highest_bid":"7","quote_volume":"8"}]`),
	}
	suite.origin.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr := suite.origin.Fetch([]Pair{pair})
	suite.NoError(cr[0].Error)
	suite.Equal(5.0, cr[0].Tick.Price)
	suite.Equal(6.0, cr[0].Tick.Ask)
	suite.Equal(7.0, cr[0].Tick.Bid)
	suite.Equal(8.0, cr[0].Tick.Volume24h)
	suite.Greater(cr[0].Tick.Timestamp.Unix(), int64(0))
}

func (suite *GateioSuite) TestRealAPICall() {
	gateio := &Gateio{Pool: query.NewHTTPWorkerPool(1)}
	testRealAPICall(suite, gateio, "ETH", "BTC")
	testRealBatchAPICall(suite, gateio, []Pair{
		{Base: "ZEC", Quote: "USDT"},
		{Base: "WIN", Quote: "USDT"},
		{Base: "BAT", Quote: "BTC"},
	})
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestGateioSuite(t *testing.T) {
	suite.Run(t, new(GateioSuite))
}
