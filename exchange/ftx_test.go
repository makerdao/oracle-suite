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
type FtxSuite struct {
	suite.Suite
	pool     query.WorkerPool
	exchange *Ftx
}

func (suite *FtxSuite) Exchange() Handler {
	return suite.exchange
}

// Setup exchange
func (suite *FtxSuite) SetupSuite() {
	suite.exchange = &Ftx{}
}

func (suite *FtxSuite) TearDownTest() {
	// cleanup created pool from prev test
	if suite.pool != nil {
		suite.pool = nil
	}
}

func (suite *FtxSuite) TestLocalPair() {
	suite.EqualValues("BTC/ETH", suite.exchange.LocalPairName(model.NewPair("BTC", "ETH")))
	suite.EqualValues("BTC/USDC", suite.exchange.LocalPairName(model.NewPair("BTC", "USDC")))
}

func (suite *FtxSuite) TestFailOnWrongInput() {
	// no pool
	_, err := suite.exchange.Call(nil, nil)
	suite.Equal(errNoPoolPassed, err)

	// empty pp
	_, err = suite.exchange.Call(newMockWorkerPool(nil), nil)
	suite.Error(err)

	// wrong pp
	_, err = suite.exchange.Call(newMockWorkerPool(nil), &model.PotentialPricePoint{})
	suite.Error(err)

	pp := newPotentialPricePoint("ftx", "BTC", "ETH")
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

	for n, r := range [][]byte{
		// invalid response
		[]byte(``),
		// invalid response
		[]byte(`{}`),
		// invalid success
		[]byte(`{"success":false}`),
		// invalid response
		[]byte(`{"success":true}`),
		// invalid response
		[]byte(`{"success":true,"result":{}}`),
		// invalid name
		[]byte(`{"success":true,"result":{"name":"SOME/ANOTHER"}}`),
		// invalid price (string)
		[]byte(`{"success":true,"result":{"name":"BTC/ETH","last":"1"}}`),
		// invalid ask (string)
		[]byte(`{"success":true,"result":{"name":"BTC/ETH","last":1,"ask":"2"}}`),
		// invalid bid (string)
		[]byte(`{"success":true,"result":{"name":"BTC/ETH","last":1,"ask":2,"bid":"3"}}`),
		// invalid volume (string)
		[]byte(`{"success":true,"result":{"name":"BTC/ETH","last":1,"ask":2,"bid":3,"quoteVolume24h":"4"}}`),
		// invalid success with normal result
		[]byte(`{"success":false,"result":{"name":"BTC/ETH","last":1,"ask":2,"bid":3,"quoteVolume24h":4}}`),
	} {
		suite.T().Run(fmt.Sprintf("Case-%d", n+1), func(t *testing.T) {
			resp = &query.HTTPResponse{Body: r}
			_, err = suite.exchange.Call(newMockWorkerPool(resp), pp)
			suite.Error(err)
		})
	}
}

func (suite *FtxSuite) TestSuccessResponse() {
	pp := newPotentialPricePoint("ftx", "BTC", "ETH")
	resp := &query.HTTPResponse{
		Body: []byte(`{"success":true,"result":{"name":"BTC/ETH","last":1,"ask":2,"bid":3,"quoteVolume24h":4}}`),
	}
	point, err := suite.exchange.Call(newMockWorkerPool(resp), pp)
	suite.NoError(err)
	suite.Equal(pp.Exchange, point.Exchange)
	suite.Equal(pp.Pair, point.Pair)
	suite.Equal(1.0, point.Price)
	suite.Equal(2.0, point.Ask)
	suite.Equal(3.0, point.Bid)
	suite.Equal(4.0, point.Volume)
	suite.Greater(point.Timestamp, int64(0))
}

func (suite *FtxSuite) TestRealAPICall() {
	testRealAPICall(suite, "ETH", "BTC")
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestFtxSuite(t *testing.T) {
	suite.Run(t, new(FtxSuite))
}
