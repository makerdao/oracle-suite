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
type KrakenSuite struct {
	suite.Suite
	pool     query.WorkerPool
	exchange *Kraken
}

func (suite *KrakenSuite) Exchange() Handler {
	return suite.exchange
}

// Setup exchange
func (suite *KrakenSuite) SetupSuite() {
	suite.exchange = &Kraken{Pool: query.NewMockWorkerPool()}
}

func (suite *KrakenSuite) TearDownTest() {
	// cleanup created pool from prev test
	if suite.pool != nil {
		suite.pool = nil
	}
}

func (suite *KrakenSuite) TestLocalPair() {
	suite.EqualValues("XXBTXETH", suite.exchange.localPairName(model.NewPair("BTC", "ETH")))
	suite.EqualValues("XXBTZUSD", suite.exchange.localPairName(model.NewPair("BTC", "USD")))
}

func (suite *KrakenSuite) TestFailOnWrongInput() {
	var err error

	// empty pp
	_, err = suite.exchange.Call([]*model.PotentialPricePoint{nil})
	suite.Error(err)

	// wrong pp
	_, err = suite.exchange.Call([]*model.PotentialPricePoint{{}})
	suite.Error(err)

	pp := newPotentialPricePoint("kraken", "DAI", "USD")
	// nil as response
	_, err = suite.exchange.Call([]*model.PotentialPricePoint{pp})
	suite.Equal(errEmptyExchangeResponse, err)

	// error in response
	ourErr := fmt.Errorf("error")
	resp := &query.HTTPResponse{
		Error: ourErr,
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	_, err = suite.exchange.Call([]*model.PotentialPricePoint{pp})
	suite.Equal(ourErr, err)

	// Error unmarshal
	resp = &query.HTTPResponse{
		Body: []byte(""),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	_, err = suite.exchange.Call([]*model.PotentialPricePoint{pp})
	suite.Error(err)

	// Error
	resp = &query.HTTPResponse{
		Body: []byte(`{"error":["abcd"]}`),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	_, err = suite.exchange.Call([]*model.PotentialPricePoint{pp})
	suite.Error(err)

	// Error
	resp = &query.HTTPResponse{
		Body: []byte(`{"error":[], "result":{}}`),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	_, err = suite.exchange.Call([]*model.PotentialPricePoint{pp})
	suite.Error(err)

	// Error
	resp = &query.HTTPResponse{
		Body: []byte(`{"error":[], "result":{"XDAIZUSD":{}}}`),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	_, err = suite.exchange.Call([]*model.PotentialPricePoint{pp})
	suite.Error(err)
}

func (suite *KrakenSuite) TestSuccessResponse() {
	pp := newPotentialPricePoint("kraken", "DAI", "USD")
	resp := &query.HTTPResponse{
		Body: []byte(`{"error":[],"result":{"DAIZUSD":{"c":["1"],"v":["2"]}}}`),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	point, err := suite.exchange.Call([]*model.PotentialPricePoint{pp})
	suite.NoError(err)
	suite.Equal(pp.Exchange, point[0].Exchange)
	suite.Equal(pp.Pair, point[0].Pair)
	suite.Equal(1.0, point[0].Price)
	suite.Equal(2.0, point[0].Volume)
	suite.Greater(point[0].Timestamp, int64(0))
}

func (suite *KrakenSuite) TestRealAPICall() {
	testRealAPICall(suite, &Kraken{Pool: query.NewHTTPWorkerPool(1)}, "ETH", "BTC")
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestKrakenSuite(t *testing.T) {
	suite.Run(t, new(KrakenSuite))
}
