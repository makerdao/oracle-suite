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
type CryptoCompareSuite struct {
	suite.Suite
	pool     query.WorkerPool
	exchange *CryptoCompare
}

func (suite *CryptoCompareSuite) Exchange() Handler {
	return suite.exchange
}

// Setup exchange
func (suite *CryptoCompareSuite) SetupSuite() {
	suite.exchange = &CryptoCompare{Pool: query.NewMockWorkerPool()}
}

func (suite *CryptoCompareSuite) TearDownTest() {
	// cleanup created pool from prev test
	if suite.pool != nil {
		suite.pool = nil
	}
}

func (suite *CryptoCompareSuite) TestFailOnWrongInput() {
	// empty pp
	cr := suite.exchange.Call([]*model.PotentialPricePoint{nil})
	suite.Len(cr, 1)
	suite.Nil(cr[0].PricePoint)
	suite.Error(cr[0].Error)

	// wrong pp
	cr = suite.exchange.Call([]*model.PotentialPricePoint{{}})
	suite.Error(cr[0].Error)

	pp := newPotentialPricePoint("cryptocompare", "BTC", "ETH")
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

	for n, r := range [][]byte{
		// invalid response
		[]byte(``),
		// invalid response
		[]byte(`{}`),
		// invalid quote
		[]byte(`{"NON":1.1}`),
		// invalid price
		[]byte(`{"ETH":"1.1"}`),
	} {
		suite.T().Run(fmt.Sprintf("Case-%d", n+1), func(t *testing.T) {
			resp = &query.HTTPResponse{Body: r}
			suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
			cr = suite.exchange.Call([]*model.PotentialPricePoint{pp})
			suite.Error(cr[0].Error)
		})
	}
}

func (suite *CryptoCompareSuite) TestSuccessResponse() {
	pp := newPotentialPricePoint("cryptocompare", "BTC", "ETH")
	resp := &query.HTTPResponse{
		Body: []byte(`{"ETH":1.1}`),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr := suite.exchange.Call([]*model.PotentialPricePoint{pp})
	suite.NoError(cr[0].Error)
	suite.Equal(pp.Exchange, cr[0].PricePoint.Exchange)
	suite.Equal(pp.Pair, cr[0].PricePoint.Pair)
	suite.Equal(1.1, cr[0].PricePoint.Price)
	suite.Greater(cr[0].PricePoint.Timestamp, int64(0))
}

func (suite *CryptoCompareSuite) TestRealAPICall() {
	testRealAPICall(suite, &CryptoCompare{Pool: query.NewHTTPWorkerPool(1)}, "ETH", "BTC")
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestCryptoCompareSuite(t *testing.T) {
	suite.Run(t, new(CryptoCompareSuite))
}
