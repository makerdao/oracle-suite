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
	"github.com/makerdao/gofer/pkg/model"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type GateioSuite struct {
	suite.Suite
	pool     query.WorkerPool
	exchange *Gateio
}

func (suite *GateioSuite) Exchange() Handler {
	return suite.exchange
}

// Setup exchange
func (suite *GateioSuite) SetupSuite() {
	suite.exchange = &Gateio{Pool: query.NewMockWorkerPool()}
}

func (suite *GateioSuite) TearDownTest() {
	// cleanup created pool from prev test
	if suite.pool != nil {
		suite.pool = nil
	}
}

func (suite *GateioSuite) TestLocalPair() {
	suite.EqualValues("BTC_ETH", suite.exchange.localPairName(model.NewPair("BTC", "ETH")))
	suite.EqualValues("BTC_USD", suite.exchange.localPairName(model.NewPair("BTC", "USD")))
}

func (suite *GateioSuite) TestFailOnWrongInput() {
	// No PPPs.
	cr := suite.exchange.Call([]*model.PotentialPricePoint{})
	suite.Len(cr, 0)

	pp := newPotentialPricePoint("gateio", "BTC", "ETH")
	// nil as response
	cr = suite.exchange.Call([]*model.PotentialPricePoint{pp})
	suite.Equal(errEmptyExchangeResponse, cr[0].Error.(*CallError).Unwrap())

	// error in response
	ourErr := fmt.Errorf("error")
	resp := &query.HTTPResponse{
		Error: ourErr,
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr = suite.exchange.Call([]*model.PotentialPricePoint{pp})
	suite.Equal(ourErr, cr[0].Error.(*CallError).Unwrap())

	// Error unmarshal
	resp = &query.HTTPResponse{
		Body: []byte(""),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr = suite.exchange.Call([]*model.PotentialPricePoint{pp})
	suite.Error(cr[0].Error)

	// Error unmarshal
	resp = &query.HTTPResponse{
		Body: []byte("[{}]"),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr = suite.exchange.Call([]*model.PotentialPricePoint{pp})
	suite.Error(cr[0].Error)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`[{"last":"abc"}]`),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr = suite.exchange.Call([]*model.PotentialPricePoint{pp})
	suite.Error(cr[0].Error)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`[{"last":"1","currency_pair":"abc"}]`),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr = suite.exchange.Call([]*model.PotentialPricePoint{pp})
	suite.Error(cr[0].Error)
}

func (suite *GateioSuite) TestSuccessResponse() {
	pp := newPotentialPricePoint("gateio", "C", "D")
	resp := &query.HTTPResponse{
		Body: []byte(`[{"currency_pair":"A_B","last":"1","lowest_ask":"2","highest_bid":"3","quote_volume":"4"},{"currency_pair":"C_D","last":"5","lowest_ask":"6","highest_bid":"7","quote_volume":"8"}]`),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr := suite.exchange.Call([]*model.PotentialPricePoint{pp})
	suite.NoError(cr[0].Error)
	suite.Equal(pp.Exchange, cr[0].PricePoint.Exchange)
	suite.Equal(pp.Pair, cr[0].PricePoint.Pair)
	suite.Equal(5.0, cr[0].PricePoint.Price)
	suite.Equal(6.0, cr[0].PricePoint.Ask)
	suite.Equal(7.0, cr[0].PricePoint.Bid)
	suite.Equal(8.0, cr[0].PricePoint.Volume)
	suite.Greater(cr[0].PricePoint.Timestamp, int64(2))
}

func (suite *GateioSuite) TestRealAPICall() {
	gateio := &Gateio{Pool: query.NewHTTPWorkerPool(1)}
	testRealAPICall(suite, gateio, "ETH", "BTC")
	testRealBatchAPICall(suite, gateio, []*model.PotentialPricePoint{
		newPotentialPricePoint("exchange", "ZEC", "USDT"),
		newPotentialPricePoint("exchange", "WIN", "USDT"),
		newPotentialPricePoint("exchange", "BAT", "BTC"),
	})
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestGateioSuite(t *testing.T) {
	suite.Run(t, new(GateioSuite))
}
