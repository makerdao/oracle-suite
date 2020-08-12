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
type CoinMarketCapSuite struct {
	suite.Suite
	pool     query.WorkerPool
	exchange Handler
}

// Setup exchange
func (suite *CoinMarketCapSuite) SetupSuite() {
	suite.exchange = &CoinMarketCap{}
}

func (suite *CoinMarketCapSuite) TearDownTest() {
	// cleanup created pool from prev test
	if suite.pool != nil {
		suite.pool = nil
	}
}

func (suite *CoinMarketCapSuite) TestLocalPair() {
	suite.EqualValues("825", suite.exchange.LocalPairName(model.NewPair("USDT", "USD")))
	suite.EqualValues("2496", suite.exchange.LocalPairName(model.NewPair("POLY", "USD")))
}

func (suite *CoinMarketCapSuite) TestFailOnWrongInput() {
	// no pool
	_, err := suite.exchange.Call(nil, nil)
	suite.Equal(errNoPoolPassed, err)

	// empty pp
	_, err = suite.exchange.Call(newMockWorkerPool(nil), nil)
	suite.Error(err)

	// wrong pp
	_, err = suite.exchange.Call(newMockWorkerPool(nil), &model.PotentialPricePoint{})
	suite.Error(err)

	pp := newPotentialPricePoint("coinmarketcap", "USDT", "USD")
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
		[]byte(`{}`),
		[]byte(`{"status":{}}`),
		// invalid error_code
		[]byte(`{"status":{"error_code":1}}`),
		// invalid error_message
		[]byte(`{"status":{"error_code":0,"error_message":"test"}}`),
		// empty data
		[]byte(`{"status":{"error_code":0,"error_message":""},"data":{}}`),
		// unexisting id
		[]byte(`{"status":{"error_code":0,"error_message":""},"data":{"non":{}}}`),
		// empty pair response
		[]byte(`{"status":{"error_code":0,"error_message":""},"data":{"825":{}}}`),
		// empty quote
		[]byte(`{"status":{"error_code":0,"error_message":""},"data":{"825":{"quote":{}}}}`),
		// wrong quote asset
		[]byte(`{"status":{"error_code":0,"error_message":""},"data":{"825":{"quote":{"NON":{}}}}}`),
		// wrong price
		[]byte(`{"status":{"error_code":0,"error_message":""},"data":{"825":{"quote":{"USD":{"price":"1"}}}}}`),
		// wrong volume
		[]byte(`{"status":{"error_code":0,"error_message":""},"data":{"825":{"quote":{"USD":{"price":1,"volume_24h":"2"}}}}}`),
	} {
		suite.T().Run(fmt.Sprintf("Case-%d", n+1), func(t *testing.T) {
			resp = &query.HTTPResponse{Body: r}
			_, err = suite.exchange.Call(newMockWorkerPool(resp), pp)
			suite.Error(err, string(resp.Body))
		})
	}
}

func (suite *CoinMarketCapSuite) TestSuccessResponse() {
	response := `
{
    "status": {
        "timestamp": "2020-08-12T11:52:08.146Z",
        "error_code": 0,
        "error_message": null,
        "elapsed": 20,
        "credit_count": 1,
        "notice": null
    },
    "data": {
        "825": {
            "id": 825,
            "name": "Tether",
            "symbol": "USDT",
            "slug": "tether",
            "num_market_pairs": 6212,
            "date_added": "2015-02-25T00:00:00.000Z",
            "tags": [
                "store-of-value",
                "stablecoin-asset-backed",
                "payments"
            ],
            "max_supply": null,
            "circulating_supply": 9998221723.19198,
            "total_supply": 10281372503.6705,
            "platform": {
                "id": 1027,
                "name": "Ethereum",
                "symbol": "ETH",
                "slug": "ethereum",
                "token_address": "0xdac17f958d2ee523a2206206994597c13d831ec7"
            },
            "is_active": 1,
            "cmc_rank": 4,
            "is_fiat": 0,
            "last_updated": "2020-08-12T11:51:31.000Z",
            "quote": {
                "USD": {
                    "price": 1.1,
                    "volume_24h": 2.1,
                    "percent_change_1h": 1.16868,
                    "percent_change_24h": 1.38225,
                    "percent_change_7d": 1.34235,
                    "market_cap": 10155531475.788067,
                    "last_updated": "2020-08-12T11:51:31.000Z"
                }
            }
        }
    }
}
`
	pp := newPotentialPricePoint("coinmarketcap", "USDT", "USD")
	resp := &query.HTTPResponse{
		Body: []byte(response),
	}
	point, err := suite.exchange.Call(newMockWorkerPool(resp), pp)
	suite.NoError(err)
	suite.Equal(pp.Exchange, point.Exchange)
	suite.Equal(pp.Pair, point.Pair)
	suite.Equal(1.1, point.Price)
	suite.Equal(2.1, point.Volume)
	suite.Greater(point.Timestamp, int64(0))
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestCoinMarketCapSuite(t *testing.T) {
	suite.Run(t, new(CoinMarketCapSuite))
}
