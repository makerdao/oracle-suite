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
type FxSuite struct {
	suite.Suite
	pool     query.WorkerPool
	exchange *Fx
}

func (suite *FxSuite) Exchange() Handler {
	return suite.exchange
}

// Setup exchange
func (suite *FxSuite) SetupSuite() {
	suite.exchange = &Fx{Pool: query.NewMockWorkerPool()}
}

func (suite *FxSuite) TearDownTest() {
	// cleanup created pool from prev test
	if suite.pool != nil {
		suite.pool = nil
	}
}

func (suite *FxSuite) TestLocalPair() {
	suite.EqualValues("BTC", suite.exchange.localPairName(model.NewPair("BTC", "ETH")))
}

func (suite *FxSuite) TestFailOnWrongInput() {
	// wrong pp
	pps := []*model.PricePoint{{}}
	suite.exchange.Fetch(pps)
	suite.Error(pps[0].Error)

	pp := newPricePoint("fx", "BTC", "ETH")
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

	// Error unmarshal
	resp = &query.HTTPResponse{
		Body: []byte(""),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	pps = []*model.PricePoint{pp}
	suite.exchange.Fetch(pps)
	suite.Error(pps[0].Error)

	// Error convert price to number
	resp = &query.HTTPResponse{
		Body: []byte(`{"rates":{}}`),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	pps = []*model.PricePoint{pp}
	suite.exchange.Fetch(pps)
	suite.Error(pps[0].Error)

	// Error convert price to number
	resp = &query.HTTPResponse{
		Body: []byte(`{"rates":{"ETH":"abcd"}}`),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	pps = []*model.PricePoint{pp}
	suite.exchange.Fetch(pps)
	suite.Error(pps[0].Error)
}

func (suite *FxSuite) TestSuccessResponse() {
	pp := newPricePoint("fx", "BTC", "ETH")
	resp := &query.HTTPResponse{
		Body: []byte(`{"rates":{"ETH":1}}`),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	pps := []*model.PricePoint{pp}
	suite.exchange.Fetch(pps)
	suite.NoError(pps[0].Error)
	suite.Equal(pp.Exchange, pps[0].Exchange)
	suite.Equal(pp.Pair, pps[0].Pair)
	suite.Equal(1.0, pps[0].Price)
	suite.Greater(pps[0].Timestamp, int64(0))
}

func (suite *FxSuite) TestRealAPICall() {
	testRealAPICall(suite, &Fx{Pool: query.NewHTTPWorkerPool(1)}, "USD", "EUR")
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestFxSuite(t *testing.T) {
	suite.Run(t, new(FxSuite))
}
