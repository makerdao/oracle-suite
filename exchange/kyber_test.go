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

	"github.com/stretchr/testify/suite"

	"github.com/makerdao/gofer/model"
	"github.com/makerdao/gofer/query"
)

type KyberSuite struct {
	suite.Suite
	pool     query.WorkerPool
	exchange Handler
}

func (suite *KyberSuite) SetupSuite() {
	suite.exchange = &Kyber{}
}

func (suite *KyberSuite) TearDownTest() {
	if suite.pool != nil {
		suite.pool = nil
	}
}

func (suite *KyberSuite) TestGetUrl() {
	suite.EqualValues("https://api.kyber.network/buy_rate?id=0x4f3afec4e5a3f2a6a1a411def7d7dfe50ee057bf&qty=2.5", suite.exchange.GetURL(newPotentialPricePoint("kyber", "DGX", "ETH")))
	suite.EqualValues("https://api.kyber.network/buy_rate?id=0xdd974d5c2e2928dea5f71b9825b8b646686bd200&qty=2.5", suite.exchange.GetURL(newPotentialPricePoint("kyber", "KNC", "ETH")))
	suite.EqualValues("https://api.kyber.network/buy_rate?id=0x80fB784B7eD66730e8b1DBd9820aFD29931aab03&qty=2.5", suite.exchange.GetURL(newPotentialPricePoint("kyber", "LEND", "ETH")))
	suite.EqualValues("https://api.kyber.network/buy_rate?id=0x9f8f72aa9304c8b593d555f12ef6589cc3a579a2&qty=2.5", suite.exchange.GetURL(newPotentialPricePoint("kyber", "MKR", "ETH")))
	suite.EqualValues("https://api.kyber.network/buy_rate?id=0x2260fac5e5542a773aa44fbcfedf7c193bc2c599&qty=2.5", suite.exchange.GetURL(newPotentialPricePoint("kyber", "WBTC", "ETH")))
}

func (suite *KyberSuite) TestFailOnWrongInput() {
	// no pool
	_, err := suite.exchange.Call(nil, nil)
	suite.Equal(errNoPoolPassed, err)

	// empty pp
	_, err = suite.exchange.Call(newMockWorkerPool(nil), nil)
	suite.Error(err)

	// wrong pp
	_, err = suite.exchange.Call(newMockWorkerPool(nil), &model.PotentialPricePoint{})
	suite.Error(err)

	pp := newPotentialPricePoint("kyber", "WBTC", "ETH")
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

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`{"data":{}`),
	}
	_, err = suite.exchange.Call(newMockWorkerPool(resp), pp)
	suite.Error(err)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`{"data":[]`),
	}
	_, err = suite.exchange.Call(newMockWorkerPool(resp), pp)
	suite.Error(err)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`{"data":[],"error":true,"reason":"yes","additional_data":"sir"}`),
	}
	_, err = suite.exchange.Call(newMockWorkerPool(resp), pp)
	suite.Error(err)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`{"data":[{"src_id":"0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee","dst_id":"0x2260fac5e5542a773aa44fbcfedf7c193bc2c599","src_qty":[],"dst_qty":[2.5]}],"error":false}`),
	}
	_, err = suite.exchange.Call(newMockWorkerPool(resp), pp)
	suite.Error(err)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`{"data":[{"src_id":"0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee","dst_id":"0x2260fac5e5542a773aa44fbcfedf7c193bc2c599","src_qty":[25.0],"dst_qty":[]}],"error":false}`),
	}
	_, err = suite.exchange.Call(newMockWorkerPool(resp), pp)
	suite.Error(err)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`{"data":[{"src_id":"0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee","dst_id":"0x2260fac5e5542a773aa44fbcfedf7c193bc2c599","src_qty":[25.0],"dst_qty":[1]}],"error":false}`),
	}
	_, err = suite.exchange.Call(newMockWorkerPool(resp), pp)
	suite.Error(err)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`{"data":[{"src_id":"0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee","dst_id":"0x2260fac5e5542a773aa44fbcfedf7c193bc2c599","src_qty":[0],"dst_qty":[2.5]}],"error":false}`),
	}
	_, err = suite.exchange.Call(newMockWorkerPool(resp), pp)
	suite.Error(err)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`{"data":[{"src_id":"0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee","dst_id":"0x2260fac5e5542a773aa44fbcfedf7c193bc2c599","src_qty":[-1.5],"dst_qty":[2.5]}],"error":false}`),
	}
	_, err = suite.exchange.Call(newMockWorkerPool(resp), pp)
	suite.Error(err)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`{"data":[{"src_id":"0xe","dst_id":"0x2260fac5e5542a773aa44fbcfedf7c193bc2c599","src_qty":[25.0],"dst_qty":[2.5]}],"error":false}`),
	}
	_, err = suite.exchange.Call(newMockWorkerPool(resp), pp)
	suite.Error(err)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`{"data":[{"src_id":"0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee","dst_id":"0xe","src_qty":[25.0],"dst_qty":[2.5]}],"error":false}`),
	}
	_, err = suite.exchange.Call(newMockWorkerPool(resp), pp)
	suite.Error(err)
}

func (suite *KyberSuite) TestSuccessResponse() {
	pp := newPotentialPricePoint("kyber", "WBTC", "ETH")
	resp := &query.HTTPResponse{
		Body: []byte(`{"data":[{"src_id":"0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee","dst_id":"0x2260fac5e5542a773aa44fbcfedf7c193bc2c599","src_qty":[25.0],"dst_qty":[2.5]}],"error":false}`),
	}
	point, err := suite.exchange.Call(newMockWorkerPool(resp), pp)
	suite.NoError(err)
	suite.Equal(pp.Exchange, point.Exchange)
	suite.Equal(pp.Pair, point.Pair)
	suite.Equal(10.0, point.Price)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestKyberSuiteSuite(t *testing.T) {
	suite.Run(t, new(KyberSuite))
}
