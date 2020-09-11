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

	"github.com/stretchr/testify/suite"

	"github.com/makerdao/gofer/internal/query"
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

func (suite *KyberSuite) TestFailOnWrongInput() {
	// wrong pair
	cr := suite.exchange.Fetch([]Pair{{}})
	suite.Error(cr[0].Error)

	pair := Pair{Base: "WBTC", Quote: "ETH"}
	// nil as response
	cr = suite.exchange.Fetch([]Pair{pair})
	suite.Equal(errEmptyExchangeResponse, cr[0].Error)

	// error in response
	ourErr := fmt.Errorf("error")
	resp := &query.HTTPResponse{
		Error: ourErr,
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr = suite.exchange.Fetch([]Pair{pair})
	suite.Equal(ourErr, cr[0].Error)

	// Error unmarshal
	resp = &query.HTTPResponse{
		Body: []byte(""),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr = suite.exchange.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)

	// Error unmarshal
	resp = &query.HTTPResponse{
		Body: []byte("{}"),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr = suite.exchange.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`{"data":{}`),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr = suite.exchange.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`{"data":[]`),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr = suite.exchange.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`{"data":[],"error":true,"reason":"yes","additional_data":"sir"}`),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr = suite.exchange.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`{"data":[{"src_id":"0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee","dst_id":"0x2260fac5e5542a773aa44fbcfedf7c193bc2c599","src_qty":[],"dst_qty":[2.5]}],"error":false}`),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr = suite.exchange.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`{"data":[{"src_id":"0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee","dst_id":"0x2260fac5e5542a773aa44fbcfedf7c193bc2c599","src_qty":[25.0],"dst_qty":[]}],"error":false}`),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr = suite.exchange.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`{"data":[{"src_id":"0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee","dst_id":"0x2260fac5e5542a773aa44fbcfedf7c193bc2c599","src_qty":[25.0],"dst_qty":[1]}],"error":false}`),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr = suite.exchange.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`{"data":[{"src_id":"0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee","dst_id":"0x2260fac5e5542a773aa44fbcfedf7c193bc2c599","src_qty":[0],"dst_qty":[2.5]}],"error":false}`),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr = suite.exchange.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`{"data":[{"src_id":"0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee","dst_id":"0x2260fac5e5542a773aa44fbcfedf7c193bc2c599","src_qty":[-1.5],"dst_qty":[2.5]}],"error":false}`),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr = suite.exchange.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`{"data":[{"src_id":"0xe","dst_id":"0x2260fac5e5542a773aa44fbcfedf7c193bc2c599","src_qty":[25.0],"dst_qty":[2.5]}],"error":false}`),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr = suite.exchange.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)

	// Error parsing
	resp = &query.HTTPResponse{
		Body: []byte(`{"data":[{"src_id":"0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee","dst_id":"0xe","src_qty":[25.0],"dst_qty":[2.5]}],"error":false}`),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr = suite.exchange.Fetch([]Pair{pair})
	suite.Error(cr[0].Error)
}

func (suite *KyberSuite) TestSuccessResponse() {
	pair := Pair{Base: "WBTC", Quote: "ETH"}
	resp := &query.HTTPResponse{
		Body: []byte(`{"data":[{"src_id":"0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee","dst_id":"0x2260fac5e5542a773aa44fbcfedf7c193bc2c599","src_qty":[25.0],"dst_qty":[2.5]}],"error":false}`),
	}
	suite.exchange.Pool.(*query.MockWorkerPool).MockResp(resp)
	cr := suite.exchange.Fetch([]Pair{pair})
	suite.NoError(cr[0].Error)
	suite.Equal(10.0, cr[0].Tick.Price)
}

func (suite *KyberSuite) TestRealAPICall() {
	testRealAPICall(suite, &Kyber{Pool: query.NewHTTPWorkerPool(1)}, "WBTC", "ETH")
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestKyberSuiteSuite(t *testing.T) {
	suite.Run(t, new(KyberSuite))
}
