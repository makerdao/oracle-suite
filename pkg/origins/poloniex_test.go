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

package origins

import (
	"fmt"
	"testing"

	"github.com/makerdao/gofer/internal/query"

	"github.com/stretchr/testify/suite"
)

type PoloniexSuite struct {
	suite.Suite
	origin *Poloniex
}

func (suite *PoloniexSuite) Origin() Handler {
	return suite.origin
}

func (suite *PoloniexSuite) SetupSuite() {
	suite.origin = &Poloniex{Pool: query.NewMockWorkerPool()}
}

func (suite *PoloniexSuite) TestLocalPair() {
	suite.EqualValues("ETH_BTC", suite.origin.localPairName(Pair{Base: "BTC", Quote: "ETH"}))
}

func (suite *PoloniexSuite) TestFailOnWrongInput() {
	pair := Pair{Base: "BTC", Quote: "ETH"}

	// Wrong pair
	cr := suite.origin.Fetch([]Pair{{}})
	suite.Error(cr[0].Error)

	// Nil as a response
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Equal(errEmptyOriginResponse, cr[0].Error)

	// Error in a response
	ourErr := fmt.Errorf("error")
	resp := &query.HTTPResponse{
		Error: ourErr,
	}

	suite.origin.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Equal(ourErr, cr[0].Error)

	// Error during unmarshalling
	resp = &query.HTTPResponse{
		Body: []byte(""),
	}
	suite.origin.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)

	// Error during converting price to a number
	resp = &query.HTTPResponse{
		Body: []byte(`
			{
				"ETH_BTC":{
					"last":"abc",
					"lowestAsk":"1.3",
					"highestBid":"1.0",
					"baseVolume":"10.1",
					"isFrozen": "0"
				}
			}
		`),
	}
	suite.origin.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)

	// Frozen pair
	resp = &query.HTTPResponse{
		Body: []byte(`
			{
				"ETH_BTC":{
					"last":"abc",
					"lowestAsk":"1.3",
					"highestBid":"1.0",
					"baseVolume":"10.1",
					"isFrozen": "1"
				}
			}
		`),
	}
	suite.origin.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)

	// Unable to find pair
	resp = &query.HTTPResponse{
		Body: []byte(`
			{
				"AAA_BBB":{
					"last":"1.1",
					"lowestAsk":"1.3",
					"highestBid":"1.0",
					"baseVolume":"10.1",
					"isFrozen": "0"
				}
			}
		`),
	}
	suite.origin.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)
}

func (suite *PoloniexSuite) TestSuccessResponse() {
	pairBTCETH := Pair{Base: "BTC", Quote: "ETH"}
	pairBTCUSD := Pair{Base: "BTC", Quote: "USD"}

	resp := &query.HTTPResponse{
		Body: []byte(`
			{
				"ETH_BTC":{
					"last":"1.1",
					"lowestAsk":"1.3",
					"highestBid":"1.0",
					"baseVolume":"10.1",
					"isFrozen": "0"
				},
				"USD_BTC":{
					"last":"2.1",
					"lowestAsk":"2.3",
					"highestBid":"2.0",
					"baseVolume":"20.1",
					"isFrozen": "0"
				},
				"EUR_BTC":{
					"last":"3.1",
					"lowestAsk":"3.3",
					"highestBid":"3.0",
					"baseVolume":"30.1",
					"isFrozen": "0"
				}
			}
		`),
	}
	suite.origin.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr := suite.origin.Fetch([]Pair{pairBTCETH, pairBTCUSD})

	suite.Len(cr, 2)

	// BTC/ETH
	suite.NoError(cr[0].Error)
	suite.Equal(pairBTCETH, cr[0].Tick.Pair)
	suite.Equal(1.1, cr[0].Tick.Price)
	suite.Equal(1.0, cr[0].Tick.Bid)
	suite.Equal(1.3, cr[0].Tick.Ask)
	suite.Equal(10.1, cr[0].Tick.Volume24h)
	suite.Greater(cr[0].Tick.Timestamp.Unix(), int64(0))

	// BTC/USD
	suite.NoError(cr[1].Error)
	suite.Equal(pairBTCUSD, cr[1].Tick.Pair)
	suite.Equal(2.1, cr[1].Tick.Price)
	suite.Equal(2.0, cr[1].Tick.Bid)
	suite.Equal(2.3, cr[1].Tick.Ask)
	suite.Equal(20.1, cr[1].Tick.Volume24h)
	suite.Greater(cr[1].Tick.Timestamp.Unix(), int64(0))
}

func (suite *PoloniexSuite) TestRealAPICall() {
	testRealAPICall(suite, &Poloniex{Pool: query.NewHTTPWorkerPool(1)}, "ETH", "BTC")
}

func TestPoloniexSuite(t *testing.T) {
	suite.Run(t, new(PoloniexSuite))
}
