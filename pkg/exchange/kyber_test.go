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

	"github.com/makerdao/gofer/internal/query"
	"github.com/makerdao/gofer/pkg/model"
)

type KyberSuite struct {
	suite.Suite
	pool     query.WorkerPool
	exchange *Kyber
}

func (suite *KyberSuite) Exchange() Handler {
	return suite.exchange
}

func (suite *KyberSuite) SetupSuite() {
	suite.exchange = &Kyber{Pool: query.NewMockWorkerPool()}
}

func (suite *KyberSuite) TearDownTest() {
	if suite.pool != nil {
		suite.pool = nil
	}
}

func (suite *KyberSuite) TestGetUrl() {
	suite.EqualValues("https://api.kyber.network/buy_rate?id=0x4f3afec4e5a3f2a6a1a411def7d7dfe50ee057bf&qty=2.5", suite.exchange.getURL(newPotentialPricePoint("kyber", "DGX", "ETH")))
	suite.EqualValues("https://api.kyber.network/buy_rate?id=0xdd974d5c2e2928dea5f71b9825b8b646686bd200&qty=2.5", suite.exchange.getURL(newPotentialPricePoint("kyber", "KNC", "ETH")))
	suite.EqualValues("https://api.kyber.network/buy_rate?id=0x80fB784B7eD66730e8b1DBd9820aFD29931aab03&qty=2.5", suite.exchange.getURL(newPotentialPricePoint("kyber", "LEND", "ETH")))
	suite.EqualValues("https://api.kyber.network/buy_rate?id=0x9f8f72aa9304c8b593d555f12ef6589cc3a579a2&qty=2.5", suite.exchange.getURL(newPotentialPricePoint("kyber", "MKR", "ETH")))
	suite.EqualValues("https://api.kyber.network/buy_rate?id=0x2260fac5e5542a773aa44fbcfedf7c193bc2c599&qty=2.5", suite.exchange.getURL(newPotentialPricePoint("kyber", "WBTC", "ETH")))
}

func (suite *KyberSuite) TestFailOnWrongInput() {
	var err error

	// empty pp
	_, err = suite.exchange.Call([]*model.PotentialPricePoint{nil})
	suite.Error(err)

	// wrong pp
	_, err = suite.exchange.Call([]*model.PotentialPricePoint{{}})
	suite.Error(err)

	pp := newPotentialPricePoint("kyber", "WBTC", "ETH")
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

	// Error unmarshal
	resp = &query.HTTPResponse{
		Body: []byte("{}"),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	_, err = suite.exchange.Call([]*model.PotentialPricePoint{pp})
	suite.Error(err)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`{"data":{}`),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	_, err = suite.exchange.Call([]*model.PotentialPricePoint{pp})
	suite.Error(err)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`{"data":[]`),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	_, err = suite.exchange.Call([]*model.PotentialPricePoint{pp})
	suite.Error(err)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`{"data":[],"error":true,"reason":"yes","additional_data":"sir"}`),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	_, err = suite.exchange.Call([]*model.PotentialPricePoint{pp})
	suite.Error(err)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`{"data":[{"src_id":"0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee","dst_id":"0x2260fac5e5542a773aa44fbcfedf7c193bc2c599","src_qty":[],"dst_qty":[2.5]}],"error":false}`),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	_, err = suite.exchange.Call([]*model.PotentialPricePoint{pp})
	suite.Error(err)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`{"data":[{"src_id":"0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee","dst_id":"0x2260fac5e5542a773aa44fbcfedf7c193bc2c599","src_qty":[25.0],"dst_qty":[]}],"error":false}`),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	_, err = suite.exchange.Call([]*model.PotentialPricePoint{pp})
	suite.Error(err)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`{"data":[{"src_id":"0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee","dst_id":"0x2260fac5e5542a773aa44fbcfedf7c193bc2c599","src_qty":[25.0],"dst_qty":[1]}],"error":false}`),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	_, err = suite.exchange.Call([]*model.PotentialPricePoint{pp})
	suite.Error(err)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`{"data":[{"src_id":"0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee","dst_id":"0x2260fac5e5542a773aa44fbcfedf7c193bc2c599","src_qty":[0],"dst_qty":[2.5]}],"error":false}`),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	_, err = suite.exchange.Call([]*model.PotentialPricePoint{pp})
	suite.Error(err)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`{"data":[{"src_id":"0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee","dst_id":"0x2260fac5e5542a773aa44fbcfedf7c193bc2c599","src_qty":[-1.5],"dst_qty":[2.5]}],"error":false}`),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	_, err = suite.exchange.Call([]*model.PotentialPricePoint{pp})
	suite.Error(err)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`{"data":[{"src_id":"0xe","dst_id":"0x2260fac5e5542a773aa44fbcfedf7c193bc2c599","src_qty":[25.0],"dst_qty":[2.5]}],"error":false}`),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	_, err = suite.exchange.Call([]*model.PotentialPricePoint{pp})
	suite.Error(err)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`{"data":[{"src_id":"0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee","dst_id":"0xe","src_qty":[25.0],"dst_qty":[2.5]}],"error":false}`),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	_, err = suite.exchange.Call([]*model.PotentialPricePoint{pp})
	suite.Error(err)
}

func (suite *KyberSuite) TestSuccessResponse() {
	pp := newPotentialPricePoint("kyber", "WBTC", "ETH")
	resp := &query.HTTPResponse{
		Body: []byte(`{"data":[{"src_id":"0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee","dst_id":"0x2260fac5e5542a773aa44fbcfedf7c193bc2c599","src_qty":[25.0],"dst_qty":[2.5]}],"error":false}`),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	point, err := suite.exchange.Call([]*model.PotentialPricePoint{pp})
	suite.NoError(err)
	suite.Equal(pp.Exchange, point[0].Exchange)
	suite.Equal(pp.Pair, point[0].Pair)
	suite.Equal(10.0, point[0].Price)
}

func (suite *KyberSuite) TestRealAPICall() {
	testRealAPICall(suite, &Kyber{Pool: query.NewHTTPWorkerPool(1)}, "WBTC", "ETH")
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestKyberSuiteSuite(t *testing.T) {
	suite.Run(t, new(KyberSuite))
}
