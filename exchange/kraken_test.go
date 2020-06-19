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
type KrakenSuite struct {
	suite.Suite
	pool     query.WorkerPool
	exchange Handler
}

// Setup exchange
func (suite *KrakenSuite) SetupSuite() {
	suite.exchange = &Kraken{}
}

func (suite *KrakenSuite) TearDownTest() {
	// cleanup created pool from prev test
	if suite.pool != nil {
		suite.pool = nil
	}
}

func (suite *KrakenSuite) TestLocalPair() {
	suite.EqualValues("XXBTZETH", suite.exchange.LocalPairName(model.NewPair("BTC", "ETH")))
	suite.EqualValues("XXBTZUSD", suite.exchange.LocalPairName(model.NewPair("BTC", "USD")))
}

func (suite *KrakenSuite) TestFailOnWrongInput() {
	// no pool
	_, err := suite.exchange.Call(nil, nil)
	suite.Equal(errNoPoolPassed, err)

	// empty pp
	_, err = suite.exchange.Call(newMockWorkerPool(nil), nil)
	suite.Error(err)

	// wrong pp
	_, err = suite.exchange.Call(newMockWorkerPool(nil), &model.PotentialPricePoint{})
	suite.Error(err)

	pp := newPotentialPricePoint("kraken", "DAI", "USD")
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

	// Error
	resp = &query.HTTPResponse{
		Body: []byte(`{"error":["abcd"]}`),
	}
	_, err = suite.exchange.Call(newMockWorkerPool(resp), pp)
	suite.Error(err)

	// Error
	resp = &query.HTTPResponse{
		Body: []byte(`{"error":[], "result":{}}`),
	}
	_, err = suite.exchange.Call(newMockWorkerPool(resp), pp)
	suite.Error(err)

	// Error
	resp = &query.HTTPResponse{
		Body: []byte(`{"error":[], "result":{"XDAIZUSD":{}}}`),
	}
	_, err = suite.exchange.Call(newMockWorkerPool(resp), pp)
	suite.Error(err)
}

func (suite *KrakenSuite) TestSuccessResponse() {
	pp := newPotentialPricePoint("kraken", "DAI", "USD")
	resp := &query.HTTPResponse{
		Body: []byte(`{"error":[],"result":{"XDAIZUSD":{"c":["1"],"v":["2"]}}}`),
	}
	point, err := suite.exchange.Call(newMockWorkerPool(resp), pp)
	suite.NoError(err)
	suite.Equal(pp.Exchange, point.Exchange)
	suite.Equal(pp.Pair, point.Pair)
	suite.Equal(1.0, point.Price)
	suite.Equal(2.0, point.Volume)
	suite.Greater(point.Timestamp, int64(0))
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestKrakenSuite(t *testing.T) {
	suite.Run(t, new(KrakenSuite))
}
