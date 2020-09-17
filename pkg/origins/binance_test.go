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

type BinanceSuite struct {
	suite.Suite
	origin *Binance
}

func (suite *BinanceSuite) Origin() Handler {
	return suite.origin
}

func (suite *BinanceSuite) SetupSuite() {
	suite.origin = &Binance{Pool: query.NewMockWorkerPool()}
}

func (suite *BinanceSuite) TestLocalPair() {
	suite.EqualValues("BTCETH", suite.origin.localPairName(Pair{Base: "BTC", Quote: "ETH"}))
	suite.NotEqual("BTCUSDC", suite.origin.localPairName(Pair{Base: "BTC", Quote: "USD"}))
}

func (suite *BinanceSuite) TestFailOnWrongInput() {
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

	// Error during during price to number
	resp = &query.HTTPResponse{
		Body: []byte(`[{"symbol":"BTCETH", "lastPrice": "abc", "bidPrice": "0", "askPrice": "0", "volume": "0", "closeTime": "10000"}]`),
	}
	suite.origin.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)

	// Unable to find pair
	resp = &query.HTTPResponse{
		Body: []byte(`[{"symbol":"AAABBB", "lastPrice": "0", "bidPrice": "0", "askPrice": "0", "volume": "0", "closeTime": "10000"}]`),
	}
	suite.origin.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr = suite.origin.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)
}

func (suite *BinanceSuite) TestSuccessResponse() {
	pairBTCETH := Pair{Base: "BTC", Quote: "ETH"}
	pairBTCUSD := Pair{Base: "BTC", Quote: "USD"}

	resp := &query.HTTPResponse{
		Body: []byte(`
			[
			   {
				  "symbol":"BTCETH",
				  "lastPrice":"1.1",
				  "bidPrice":"1.0",
				  "askPrice":"1.3",
				  "volume":"10.1",
				  "closeTime":10000
			   },
			   {
				  "symbol":"BTCUSD",
				  "lastPrice":"2.1",
				  "bidPrice":"2.0",
				  "askPrice":"2.3",
				  "volume":"20.1",
				  "closeTime":10000
			   },
			   {
				  "symbol":"BTCEUR",
				  "lastPrice":"3.1",
				  "bidPrice":"3.0",
				  "askPrice":"3.3",
				  "volume":"30.1",
				  "closeTime":10000
			   }
			]
		`),
	}
	suite.origin.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr := suite.origin.Fetch([]Pair{pairBTCETH, pairBTCUSD})

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

func (suite *BinanceSuite) TestRealAPICall() {
	testRealAPICall(suite, &Binance{Pool: query.NewHTTPWorkerPool(1)}, "ETH", "BTC")
}

func TestBinanceSuite(t *testing.T) {
	suite.Run(t, new(BinanceSuite))
}
