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
	"encoding/json"
	"testing"

	"github.com/makerdao/oracle-suite/internal/query"

	"github.com/stretchr/testify/suite"
)

type CurveSuite struct {
	suite.Suite
	addresses ContractAddresses
	pool      *query.MockWorkerPool
	origin    *BaseExchangeHandler
}

func (suite *CurveSuite) SetupSuite() {
	suite.addresses = ContractAddresses{
		"ETH/STETH": "0xDC24316b9AE028F1497c275EB9192a3Ea0f67022",
	}
	suite.pool = query.NewMockWorkerPool()
}
func (suite *CurveSuite) TearDownSuite() {
	suite.addresses = nil
	suite.pool = nil
}

func (suite *CurveSuite) SetupTest() {
	curveFinance, err := NewCurveFinance("", suite.pool, suite.addresses)
	suite.NoError(err)
	suite.origin = NewBaseExchangeHandler(curveFinance, nil)
}

func (suite *CurveSuite) TearDownTest() {
	suite.origin = nil
}

func (suite *CurveSuite) Origin() Handler {
	return suite.origin
}

func TestCurveSuite(t *testing.T) {
	suite.Run(t, new(CurveSuite))
}

func (suite *CurveSuite) TestSuccessResponse() {
	suite.pool.MockBody(`{"jsonrpc":"2.0","id":1,"result":"0x0000000000000000000000000000000000000000000000000dc19f91822f3fe3"}`)
	suite.pool.SetRequestAssertions(func(req *query.HTTPRequest) {
		suite.NotEmpty(req)
		suite.NotEmpty(req.Body)

		var request jsonrpcMessage
		err := json.NewDecoder(req.Body).Decode(&request)
		suite.Require().NoError(err)

		suite.True(request.isCall())
		suite.Equal("eth_call", request.Method)

		suite.Contains(string(request.Params), `{"to":"0xDC24316b9AE028F1497c275EB9192a3Ea0f67022","data":"0x5e0d443f`)
	})

	pair := Pair{Base: "STETH", Quote: "ETH"}

	results1 := suite.origin.Fetch([]Pair{pair})
	suite.Require().NoError(results1[0].Error)
	suite.Equal(0.9912488403014287, results1[0].Price.Price)
	suite.Greater(results1[0].Price.Timestamp.Unix(), int64(0))

	results2 := suite.origin.Fetch([]Pair{pair.Inverse()})
	suite.Require().NoError(results2[0].Error)
	suite.Equal(0.9912488403014287, results2[0].Price.Price)
	suite.Greater(results2[0].Price.Timestamp.Unix(), int64(0))
}

func (suite *CurveSuite) TestSuccessResponse2() {
	suite.pool.MockBody(`{"jsonrpc":"2.0","id":1,"result":"0x0000000000000000000000000000000000000000000000000dc19f91822f3fe3"}`)
	suite.pool.SetRequestAssertions(func(req *query.HTTPRequest) {
		suite.NotEmpty(req)
		suite.NotEmpty(req.Body)

		var request jsonrpcMessage
		err := json.NewDecoder(req.Body).Decode(&request)
		suite.Require().NoError(err)

		suite.True(request.isCall())
		suite.Equal("eth_call", request.Method)

		suite.Contains(string(request.Params), `{"to":"0xDC24316b9AE028F1497c275EB9192a3Ea0f67022","data":"0x5e0d443f`)
	})

	pair := Pair{Base: "ETH", Quote: "STETH"}
	cr := suite.origin.Fetch([]Pair{pair})
	suite.Require().NoError(cr[0].Error)
	suite.Equal(0.9912488403014287, cr[0].Price.Price)
	suite.Greater(cr[0].Price.Timestamp.Unix(), int64(0))
}

func (suite *CurveSuite) TestFailOnWrongPair() {
	pair := Pair{Base: "x", Quote: "y"}
	cr := suite.origin.Fetch([]Pair{pair})
	suite.Require().EqualError(cr[0].Error, "failed to get contract address for pair: x/y")
}
