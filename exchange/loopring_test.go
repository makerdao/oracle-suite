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

const successResponse = `{
    "resultInfo": {
        "code": 0,
        "message": "SUCCESS"
    },
    "data": {
        "USDC-USDT": {
            "last": "0.9997",
            "lowestAsk": "1.0000",
            "highestBid": "0.9998",
            "percentChange": "0.0001",
            "baseVolume": "36061.29",
            "quoteVolume": "36053.99",
            "high24hr": "1.0002",
            "low24hr": "0.9990",
            "isFrozen": "0"
        },
        "LRC-USDT": {
            "last": "0.1221",
            "lowestAsk": "0.1221",
            "highestBid": "0.1215",
            "percentChange": "0.1192",
            "baseVolume": "1488717.990",
            "quoteVolume": "181251.50",
            "high24hr": "0.1323",
            "low24hr": "0.1091",
            "isFrozen": "0"
        }
    }
}`

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type LoopringSuite struct {
	suite.Suite
	pool     query.WorkerPool
	exchange Handler
}

// Setup exchange
func (suite *LoopringSuite) SetupSuite() {
	suite.exchange = &Loopring{}
}

func (suite *LoopringSuite) TearDownTest() {
	// cleanup created pool from prev test
	if suite.pool != nil {
		suite.pool = nil
	}
}

func (suite *LoopringSuite) TestLocalPair() {
	suite.EqualValues("USDT-DAI", suite.exchange.LocalPairName(model.NewPair("USDT", "DAI")))
	suite.EqualValues("ETH-DAI", suite.exchange.LocalPairName(model.NewPair("ETH", "DAI")))
}

func (suite *LoopringSuite) TestFailOnWrongInput() {
	// no pool
	_, err := suite.exchange.Call(nil, nil)
	suite.Equal(errNoPoolPassed, err)

	// empty pp
	_, err = suite.exchange.Call(newMockWorkerPool(nil), nil)
	suite.Error(err)

	// wrong pp
	_, err = suite.exchange.Call(newMockWorkerPool(nil), &model.PotentialPricePoint{})
	suite.Error(err)

	pp := newPotentialPricePoint("loopring", "LRC", "USDT")
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

	// Error unmarshal
	resp = &query.HTTPResponse{
		Body: []byte("{}"),
	}
	_, err = suite.exchange.Call(newMockWorkerPool(resp), pp)
	suite.Error(err)

	// Error wrong code
	resp = &query.HTTPResponse{
		Body: []byte(`{"resultInfo":{"code":1,"message":"SUCCESS"}}`),
	}
	_, err = suite.exchange.Call(newMockWorkerPool(resp), pp)
	suite.Error(err)

	// Error wrong message
	resp = &query.HTTPResponse{
		Body: []byte(`{"resultInfo":{"code":0,"message":"Wrong"}}`),
	}
	_, err = suite.exchange.Call(newMockWorkerPool(resp), pp)
	suite.Error(err)

	// Error no data
	resp = &query.HTTPResponse{
		Body: []byte(`{"resultInfo":{"code":0,"message":"SUCCESS"}}`),
	}
	_, err = suite.exchange.Call(newMockWorkerPool(resp), pp)
	suite.Error(err)
	// Error no pair in data
	resp = &query.HTTPResponse{
		Body: []byte(`{"resultInfo":{"code":0,"message":"SUCCESS"},"data":{}}`),
	}
	_, err = suite.exchange.Call(newMockWorkerPool(resp), pp)
	suite.Error(err)
}

func (suite *LoopringSuite) TestSuccessResponse() {
	pp := newPotentialPricePoint("loopring", "LRC", "USDT")
	resp := &query.HTTPResponse{
		Body: []byte(successResponse),
	}
	point, err := suite.exchange.Call(newMockWorkerPool(resp), pp)
	suite.NoError(err)
	suite.Equal(pp.Exchange, point.Exchange)
	suite.Equal(pp.Pair, point.Pair)
	suite.Equal(0.1221, point.Price)
	suite.Equal(181251.50, point.Volume)
	suite.Equal(0.1221, point.Ask)
	suite.Equal(0.1215, point.Bid)
	suite.Greater(point.Timestamp, int64(2))
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestLoopringSuite(t *testing.T) {
	suite.Run(t, new(LoopringSuite))
}
