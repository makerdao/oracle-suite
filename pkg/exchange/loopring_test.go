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
	exchange *Loopring
}

func (suite *LoopringSuite) Exchange() Handler {
	return suite.exchange
}

// Setup exchange
func (suite *LoopringSuite) SetupSuite() {
	suite.exchange = &Loopring{Pool: query.NewMockWorkerPool()}
}

func (suite *LoopringSuite) TearDownTest() {
	// cleanup created pool from prev test
	if suite.pool != nil {
		suite.pool = nil
	}
}

func (suite *LoopringSuite) TestLocalPair() {
	suite.EqualValues("USDT-DAI", suite.exchange.localPairName(model.NewPair("USDT", "DAI")))
	suite.EqualValues("ETH-DAI", suite.exchange.localPairName(model.NewPair("ETH", "DAI")))
}

func (suite *LoopringSuite) TestFailOnWrongInput() {
	// empty pp
	cr := suite.exchange.Call([]*model.PotentialPricePoint{nil})
	suite.Len(cr, 1)
	suite.Nil(cr[0].PricePoint)
	suite.Error(cr[0].Error)

	// wrong pp
	cr = suite.exchange.Call([]*model.PotentialPricePoint{{}})
	suite.Error(cr[0].Error)

	pp := newPotentialPricePoint("loopring", "LRC", "USDT")
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

	// Error unmarshal
	resp = &query.HTTPResponse{
		Body: []byte("{}"),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr = suite.exchange.Call([]*model.PotentialPricePoint{pp})
	suite.Error(cr[0].Error)

	// Error wrong code
	resp = &query.HTTPResponse{
		Body: []byte(`{"resultInfo":{"code":1,"message":"SUCCESS"}}`),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr = suite.exchange.Call([]*model.PotentialPricePoint{pp})
	suite.Error(cr[0].Error)

	// Error wrong message
	resp = &query.HTTPResponse{
		Body: []byte(`{"resultInfo":{"code":0,"message":"Wrong"}}`),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr = suite.exchange.Call([]*model.PotentialPricePoint{pp})
	suite.Error(cr[0].Error)

	// Error no data
	resp = &query.HTTPResponse{
		Body: []byte(`{"resultInfo":{"code":0,"message":"SUCCESS"}}`),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr = suite.exchange.Call([]*model.PotentialPricePoint{pp})
	suite.Error(cr[0].Error)
	// Error no pair in data
	resp = &query.HTTPResponse{
		Body: []byte(`{"resultInfo":{"code":0,"message":"SUCCESS"},"data":{}}`),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr = suite.exchange.Call([]*model.PotentialPricePoint{pp})
	suite.Error(cr[0].Error)
}

func (suite *LoopringSuite) TestSuccessResponse() {
	pp := newPotentialPricePoint("loopring", "LRC", "USDT")
	resp := &query.HTTPResponse{
		Body: []byte(successResponse),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr := suite.exchange.Call([]*model.PotentialPricePoint{pp})
	suite.NoError(cr[0].Error)
	suite.Equal(pp.Exchange, cr[0].PricePoint.Exchange)
	suite.Equal(pp.Pair, cr[0].PricePoint.Pair)
	suite.Equal(0.1221, cr[0].PricePoint.Price)
	suite.Equal(181251.50, cr[0].PricePoint.Volume)
	suite.Equal(0.1221, cr[0].PricePoint.Ask)
	suite.Equal(0.1215, cr[0].PricePoint.Bid)
	suite.Greater(cr[0].PricePoint.Timestamp, int64(2))
}

func (suite *LoopringSuite) TestRealAPICall() {
	testRealAPICall(suite, &Loopring{Pool: query.NewHTTPWorkerPool(1)}, "LRC", "ETH")
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestLoopringSuite(t *testing.T) {
	suite.Run(t, new(LoopringSuite))
}
