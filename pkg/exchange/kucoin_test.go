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
type KucoinSuite struct {
	suite.Suite
	pool     query.WorkerPool
	exchange *Kucoin
}

func (suite *KucoinSuite) Exchange() Handler {
	return suite.exchange
}

// Setup exchange
func (suite *KucoinSuite) SetupSuite() {
	suite.exchange = &Kucoin{Pool: query.NewMockWorkerPool()}
}

func (suite *KucoinSuite) TearDownTest() {
	// cleanup created pool from prev test
	if suite.pool != nil {
		suite.pool = nil
	}
}

func (suite *KucoinSuite) TestLocalPair() {
	suite.EqualValues("BTC-ETH", suite.exchange.localPairName(model.NewPair("BTC", "ETH")))
	suite.NotEqual("BTC-USDC", suite.exchange.localPairName(model.NewPair("BTC", "USD")))
}

func (suite *KucoinSuite) TestFailOnWrongInput() {
	// empty pp
	cr := suite.exchange.Call([]*model.PotentialPricePoint{nil})
	suite.Len(cr, 1)
	suite.Nil(cr[0].PricePoint)
	suite.Error(cr[0].Error)

	// wrong pp
	cr = suite.exchange.Call([]*model.PotentialPricePoint{{}})
	suite.Error(cr[0].Error)

	pp := newPotentialPricePoint("kucoin", "BTC", "ETH")
	// nil as response
	cr = suite.exchange.Call([]*model.PotentialPricePoint{pp})
	suite.Equal(errEmptyExchangeResponse, cr[0].Error)

	// error in response
	ourErr := fmt.Errorf("error")
	resp := &query.HTTPResponse{
		Error: ourErr,
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr = suite.exchange.Call([]*model.PotentialPricePoint{pp})
	suite.Equal(ourErr, cr[0].Error)

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
			"code":"200000",
			"data": {
				"time":1596632420791,
				"sequence":"1594320230985",
				"price":"err",
				"size":"0.129",
				"bestBid":"139.55",
				"bestBidSize": "0.2866",
				"bestAsk":"139.7",
				"bestAskSize":"0.2863"
			}
		}`),
		// invalid bid price
		[]byte(`{
			"code":"200000",
			"data": {
				"time":1596632420791,
				"sequence":"1594320230985",
				"price":"1.23",
				"size":"0.129",
				"bestBid":"err",
				"bestBidSize": "0.2866",
				"bestAsk":"139.7",
				"bestAskSize":"0.2863"
			}
		}`),
		// invalid ask price
		[]byte(`{
			"code":"200000",
			"data": {
				"time":1596632420791,
				"sequence":"1594320230985",
				"price":"1.23",
				"size":"0.129",
				"bestBid":"139.55",
				"bestBidSize": "0.2866",
				"bestAsk":"err",
				"bestAskSize":"0.2863"
			}
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

func (suite *KucoinSuite) TestSuccessResponse() {
	pp := newPotentialPricePoint("kucoin", "BTC", "ETH")
	resp := &query.HTTPResponse{
		Body: []byte(`{
			"code":"200000",
			"data": {
				"time":1596632420791,
				"sequence":"1594320230985",
				"price":"1.23",
				"size":"0.123",
				"bestBid":"1.2",
				"bestBidSize": "0.2866",
				"bestAsk":"1.3",
				"bestAskSize":"0.2863"
			}
		}`),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr := suite.exchange.Call([]*model.PotentialPricePoint{pp})
	suite.NoError(cr[0].Error)
	suite.Equal(pp.Exchange, cr[0].PricePoint.Exchange)
	suite.Equal(pp.Pair, cr[0].PricePoint.Pair)
	suite.Equal(int64(1596632420), cr[0].PricePoint.Timestamp)
	suite.Equal(1.23, cr[0].PricePoint.Price)
	suite.Equal(1.3, cr[0].PricePoint.Bid)
	suite.Equal(1.2, cr[0].PricePoint.Ask)
}

func (suite *KucoinSuite) TestRealAPICall() {
	testRealAPICall(suite, &Kucoin{Pool: query.NewHTTPWorkerPool(1)}, "ETH", "BTC")
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestKucoinSuite(t *testing.T) {
	suite.Run(t, new(KucoinSuite))
}
