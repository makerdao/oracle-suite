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
	suite.exchange = &Ftx{Pool: query.NewMockWorkerPool()}
}

func (suite *FtxSuite) TearDownTest() {
	// cleanup created pool from prev test
	if suite.pool != nil {
		suite.pool = nil
	}
}

func (suite *FtxSuite) TestLocalPair() {
	suite.EqualValues("BTC/ETH", suite.exchange.localPairName(model.NewPair("BTC", "ETH")))
	suite.EqualValues("BTC/USDC", suite.exchange.localPairName(model.NewPair("BTC", "USDC")))
}

func (suite *FtxSuite) TestFailOnWrongInput() {
	// wrong pp
	pps := []*model.PricePoint{{}}
	suite.exchange.Fetch(pps)
	suite.Error(pps[0].Error)

	pp := newPricePoint("ftx", "BTC", "ETH")
	// nil as response
	pps = []*model.PricePoint{pp}
	suite.exchange.Fetch(pps)
	suite.Equal(errEmptyExchangeResponse, pps[0].Error)

	// error in response
	ourErr := fmt.Errorf("error")
	resp := &query.HTTPResponse{
		Error: ourErr,
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	pps = []*model.PricePoint{pp}
	suite.exchange.Fetch(pps)
	suite.Equal(ourErr, pps[0].Error)

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
			suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
			pps = []*model.PricePoint{pp}
			suite.exchange.Fetch(pps)
			suite.Error(pps[0].Error)
		})
	}
}

func (suite *FtxSuite) TestSuccessResponse() {
	pp := newPricePoint("ftx", "BTC", "ETH")
	resp := &query.HTTPResponse{
		Body: []byte(`{"success":true,"result":{"name":"BTC/ETH","last":1,"ask":2,"bid":3,"quoteVolume24h":4}}`),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	pps := []*model.PricePoint{pp}
	suite.exchange.Fetch(pps)
	suite.NoError(pps[0].Error)
	suite.Equal(pp.Exchange, pps[0].Exchange)
	suite.Equal(pp.Pair, pps[0].Pair)
	suite.Equal(1.0, pps[0].Price)
	suite.Equal(2.0, pps[0].Ask)
	suite.Equal(3.0, pps[0].Bid)
	suite.Equal(4.0, pps[0].Volume)
	suite.Greater(pps[0].Timestamp, int64(0))
}

func (suite *FtxSuite) TestRealAPICall() {
	testRealAPICall(suite, &Ftx{Pool: query.NewHTTPWorkerPool(1)}, "ETH", "BTC")
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestFtxSuite(t *testing.T) {
	suite.Run(t, new(FtxSuite))
}
