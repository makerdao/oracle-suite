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

type BittrexSuite struct {
	suite.Suite
	origin *Bittrex
}

func (suite *BittrexSuite) Origin() Handler {
	return suite.origin
}

func (suite *BittrexSuite) SetupSuite() {
	suite.origin = &Bittrex{Pool: query.NewMockWorkerPool()}
}

func (suite *BittrexSuite) TestLocalPair() {
	suite.EqualValues("ETH-BTC", suite.origin.localPairName(Pair{Base: "BTC", Quote: "ETH"}))
}

func (suite *BittrexSuite) TestFailOnWrongInput() {
	pair := Pair{Base: "BTC", Quote: "ETH"}

	// Wrong pair
	fr := suite.origin.Fetch([]Pair{{}})
	suite.Error(fr[0].Error)

	// Nil as a response
	fr = suite.origin.Fetch([]Pair{pair})
	suite.Equal(errEmptyOriginResponse, fr[0].Error)

	// Error in a response
	ourErr := fmt.Errorf("error")
	resp := &query.HTTPResponse{
		Error: ourErr,
	}

	suite.origin.Pool.(*query.MockWorkerPool).MockResp(resp)
	fr = suite.origin.Fetch([]Pair{pair})
	suite.Equal(ourErr, fr[0].Error)

	// Error during unmarshalling
	resp = &query.HTTPResponse{
		Body: []byte(""),
	}
	suite.origin.Pool.(*query.MockWorkerPool).MockResp(resp)
	fr = suite.origin.Fetch([]Pair{pair})
	suite.Error(fr[0].Error)

	// Price as string
	resp = &query.HTTPResponse{
		Body: []byte(`
			{
			   "success":true,
			   "message":"",
			   "result":[
				  {
					 "MarketName":"BTC-ETH",
					 "Volume":10.1,
					 "Last":"1.1",
					 "TimeStamp":"2020-09-18T12:10:59.29",
					 "Bid":1.0,
					 "Ask":1.3
				  },
			   ]
			}
		`),
	}
	suite.origin.Pool.(*query.MockWorkerPool).MockResp(resp)
	fr = suite.origin.Fetch([]Pair{pair})
	suite.Error(fr[0].Error)

	// Unable to find pair
	resp = &query.HTTPResponse{
		Body: []byte(`
			{
			   "success":true,
			   "message":"",
			   "result":[
				  {
					 "MarketName":"AAA-BBB",
					 "Volume":10.1,
					 "Last":"1.1",
					 "TimeStamp":"2020-09-18T12:10:59.29",
					 "Bid":1.0,
					 "Ask":1.3
				  },
			   ]
			}
		`),
	}
	suite.origin.Pool.(*query.MockWorkerPool).MockResp(resp)
	fr = suite.origin.Fetch([]Pair{pair})
	suite.Error(fr[0].Error)
}

func (suite *BittrexSuite) TestSuccessResponse() {
	pairBTCETH := Pair{Base: "BTC", Quote: "ETH"}
	pairBTCUSD := Pair{Base: "BTC", Quote: "USD"}

	resp := &query.HTTPResponse{
		Body: []byte(`
			{
			   "success":true,
			   "message":"",
			   "result":[
				  {
					 "MarketName":"ETH-BTC",
					 "Volume":10.1,
					 "Last":1.1,
					 "TimeStamp":"2020-09-18T12:10:59.29",
					 "Bid":1.0,
					 "Ask":1.3
				  },
				  {
					 "MarketName":"USD-BTC",
					 "Volume":20.1,
					 "Last":2.1,
					 "TimeStamp":"2020-09-18T12:10:59.29",
					 "Bid":2.0,
					 "Ask":2.3
				  },
				  {
					 "MarketName":"EUR-BTC",
					 "Volume":30.1,
					 "Last":3.1,
					 "TimeStamp":"2020-09-18T12:10:59.29",
					 "Bid":3.0,
					 "Ask":3.3
				  }
			   ]
			}
		`),
	}
	suite.origin.Pool.(*query.MockWorkerPool).MockResp(resp)
	fr := suite.origin.Fetch([]Pair{pairBTCETH, pairBTCUSD})

	suite.Len(fr, 2)

	// BTC/ETH
	suite.NoError(fr[0].Error)
	suite.Equal(pairBTCETH, fr[0].Tick.Pair)
	suite.Equal(1.1, fr[0].Tick.Price)
	suite.Equal(1.0, fr[0].Tick.Bid)
	suite.Equal(1.3, fr[0].Tick.Ask)
	suite.Equal(10.1, fr[0].Tick.Volume24h)
	suite.Greater(fr[0].Tick.Timestamp.Unix(), int64(0))

	// BTC/USD
	suite.NoError(fr[1].Error)
	suite.Equal(pairBTCUSD, fr[1].Tick.Pair)
	suite.Equal(2.1, fr[1].Tick.Price)
	suite.Equal(2.0, fr[1].Tick.Bid)
	suite.Equal(2.3, fr[1].Tick.Ask)
	suite.Equal(20.1, fr[1].Tick.Volume24h)
	suite.Greater(fr[1].Tick.Timestamp.Unix(), int64(0))
}

func (suite *BittrexSuite) TestRealAPICall() {
	testRealBatchAPICall(
		suite,
		&Bittrex{Pool: query.NewHTTPWorkerPool(1)},
		[]Pair{
			{Base: "ETH", Quote: "BTC"},
			{Base: "DGX", Quote: "USDT"},
			{Base: "MKR", Quote: "ETH"},
			{Base: "OMG", Quote: "USDT"},
			{Base: "USDT", Quote: "USD"},
			{Base: "ZRX", Quote: "USD"},
		},
	)
}

func TestBittrexSuite(t *testing.T) {
	suite.Run(t, new(BittrexSuite))
}
