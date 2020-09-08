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
type OkexSuite struct {
	suite.Suite
	pool     query.WorkerPool
	exchange *Okex
}

func (suite *OkexSuite) Exchange() Handler {
	return suite.exchange
}

// Setup exchange
func (suite *OkexSuite) SetupSuite() {
	suite.exchange = &Okex{Pool: query.NewMockWorkerPool()}
}

func (suite *OkexSuite) TearDownTest() {
	// cleanup created pool from prev test
	if suite.pool != nil {
		suite.pool = nil
	}
}

func (suite *OkexSuite) TestLocalPair() {
	suite.EqualValues("BTC-ETH", suite.exchange.localPairName(model.NewPair("BTC", "ETH")))
	suite.NotEqual("BTC-USDC", suite.exchange.localPairName(model.NewPair("BTC", "USD")))
}

func (suite *OkexSuite) TestFailOnWrongInput() {
	// empty pp
	cr := suite.exchange.Call([]*model.PotentialPricePoint{nil})
	suite.Len(cr, 1)
	suite.Nil(cr[0].PricePoint)
	suite.Error(cr[0].Error)

	// wrong pp
	cr = suite.exchange.Call([]*model.PotentialPricePoint{{}})
	suite.Error(cr[0].Error)

	pp := newPotentialPricePoint("okex", "BTC", "ETH")
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

	for n, r := range [][]byte{
		// invalid price
		[]byte(`{
			"best_ask":"10000.5",
			"best_bid":"10000.4",
			"instrument_id":"BTC-USDT",
			"product_id":"BTC-USDT",
			"last":"err",
			"last_qty":"1.23",
			"ask":"10000.5",
			"best_ask_size":"1.23456789",
			"bid":"10000.4",
			"best_bid_size":"12.3456789",
			"open_24h":"10000.1",
			"high_24h":"11000.1",
			"low_24h":"9000.1",
			"base_volume_24h":"50000.1234",
			"timestamp":"2020-08-06T10:02:46.360Z",
			"quote_volume_24h":"600000001"
		}`),
		// invalid volume
		[]byte(`{
			"best_ask":"10000.5",
			"best_bid":"10000.4",
			"instrument_id":"BTC-USDT",
			"product_id":"BTC-USDT",
			"last":"10000.4",
			"last_qty":"1.23",
			"ask":"10000.5",
			"best_ask_size":"1.23456789",
			"bid":"10000.4",
			"best_bid_size":"12.3456789",
			"open_24h":"10000.1",
			"high_24h":"11000.1",
			"low_24h":"9000.1",
			"base_volume_24h":"err",
			"timestamp":"2020-08-06T10:02:46.360Z",
			"quote_volume_24h":"600000001"
		}`),
		// invalid bid price
		[]byte(`{
			"best_ask":"10000.5",
			"best_bid":"10000.4",
			"instrument_id":"BTC-USDT",
			"product_id":"BTC-USDT",
			"last":"10000.4",
			"last_qty":"1.23",
			"ask":"10000.5",
			"best_ask_size":"1.23456789",
			"bid":"err",
			"best_bid_size":"12.3456789",
			"open_24h":"10000.1",
			"high_24h":"11000.1",
			"low_24h":"9000.1",
			"base_volume_24h":"50000.1234",
			"timestamp":"2020-08-06T10:02:46.360Z",
			"quote_volume_24h":"600000001"
		}`),
		// invalid ask price
		[]byte(`{
			"best_ask":"10000.5",
			"best_bid":"10000.4",
			"instrument_id":"BTC-USDT",
			"product_id":"BTC-USDT",
			"last":"10000.4",
			"last_qty":"1.23",
			"ask":"err",
			"best_ask_size":"1.23456789",
			"bid":"10000.4",
			"best_bid_size":"12.3456789",
			"open_24h":"10000.1",
			"high_24h":"11000.1",
			"low_24h":"9000.1",
			"base_volume_24h":"50000.1234",
			"timestamp":"2020-08-06T10:02:46.360Z",
			"quote_volume_24h":"600000001"
		}`),
	} {
		suite.T().Run(fmt.Sprintf("Case-%d", n+1), func(t *testing.T) {
			resp = &query.HTTPResponse{Body: r}
			suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
			cr = suite.exchange.Call([]*model.PotentialPricePoint{pp})
			suite.Error(cr[0].Error)
		})
	}
}

func (suite *OkexSuite) TestSuccessResponse() {
	pp := newPotentialPricePoint("okex", "BTC", "ETH")
	resp := &query.HTTPResponse{
		Body: []byte(`{
			"best_ask":"10000.5",
			"best_bid":"10000.4",
			"instrument_id":"BTC-USDT",
			"product_id":"BTC-USDT",
			"last":"10000.4",
			"last_qty":"1.23",
			"ask":"10000.5",
			"best_ask_size":"1.23456789",
			"bid":"10000.4",
			"best_bid_size":"12.3456789",
			"open_24h":"10000.1",
			"high_24h":"11000.1",
			"low_24h":"9000.1",
			"base_volume_24h":"50000.1234",
			"timestamp":"2020-08-06T10:02:46.360Z",
			"quote_volume_24h":"600000001"
		}`),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr := suite.exchange.Call([]*model.PotentialPricePoint{pp})
	suite.NoError(cr[0].Error)
	suite.Equal(pp.Exchange, cr[0].PricePoint.Exchange)
	suite.Equal(pp.Pair, cr[0].PricePoint.Pair)
	suite.Equal(int64(1596708166), cr[0].PricePoint.Timestamp)
	suite.Equal(10000.4, cr[0].PricePoint.Price)
	suite.Equal(50000.1234, cr[0].PricePoint.Volume)
	suite.Equal(10000.4, cr[0].PricePoint.Bid)
	suite.Equal(10000.5, cr[0].PricePoint.Ask)
}

func (suite *OkexSuite) TestRealAPICall() {
	testRealAPICall(suite, &Okex{Pool: query.NewHTTPWorkerPool(1)}, "ETH", "BTC")
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestOkexSuite(t *testing.T) {
	suite.Run(t, new(OkexSuite))
}
