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

	"github.com/makerdao/gofer/model"
	"github.com/makerdao/gofer/query"

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
	suite.exchange = &Kucoin{}
}

func (suite *KucoinSuite) TearDownTest() {
	// cleanup created pool from prev test
	if suite.pool != nil {
		suite.pool = nil
	}
}

func (suite *KucoinSuite) TestLocalPair() {
	suite.EqualValues("BTC-ETH", suite.exchange.LocalPairName(model.NewPair("BTC", "ETH")))
	suite.NotEqual("BTC-USDC", suite.exchange.LocalPairName(model.NewPair("BTC", "USD")))
}

func (suite *KucoinSuite) TestFailOnWrongInput() {
	// no pool
	_, err := suite.exchange.Call(nil, nil)
	suite.Equal(errNoPoolPassed, err)

	// empty pp
	_, err = suite.exchange.Call(newMockWorkerPool(nil), nil)
	suite.Error(err)

	// wrong pp
	_, err = suite.exchange.Call(newMockWorkerPool(nil), &model.PotentialPricePoint{})
	suite.Error(err)

	pp := newPotentialPricePoint("kucoin", "BTC", "ETH")
	// nil as response
	_, err = suite.exchange.Call(newMockWorkerPool(nil), pp)
	suite.Equal(errEmptyExchangeResponse, err)

	// error in response
	ourErr := fmt.Errorf("error")
	resp := &query.HTTPResponse{
		Error: ourErr,
	}
	_, err = suite.exchange.Call(newMockWorkerPool(resp), pp)
	suite.Equal(ourErr, err)

	// Error unmarshal
	resp = &query.HTTPResponse{
		Body: []byte(""),
	}
	_, err = suite.exchange.Call(newMockWorkerPool(resp), pp)
	suite.Error(err)

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
			_, err = suite.exchange.Call(newMockWorkerPool(resp), pp)
			suite.Error(err)
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
	point, err := suite.exchange.Call(newMockWorkerPool(resp), pp)
	suite.NoError(err)
	suite.Equal(pp.Exchange, point.Exchange)
	suite.Equal(pp.Pair, point.Pair)
	suite.Equal(int64(1596632420), point.Timestamp)
	suite.Equal(1.23, point.Price)
	suite.Equal(1.3, point.Bid)
	suite.Equal(1.2, point.Ask)
}

func (suite *KucoinSuite) TestRealAPICall() {
	testRealAPICall(suite, "ETH", "BTC")
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestKucoinSuite(t *testing.T) {
	suite.Run(t, new(KucoinSuite))
}
