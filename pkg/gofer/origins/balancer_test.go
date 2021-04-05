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

	"github.com/makerdao/oracle-suite/internal/query"

	"github.com/stretchr/testify/suite"
)

type BalancerSuite struct {
	suite.Suite
	origin *Balancer
}

func (suite *BalancerSuite) Origin() Handler {
	return suite.origin
}

func (suite *BalancerSuite) SetupSuite() {
	suite.origin = &Balancer{Pool: query.NewMockWorkerPool()}
}

func (suite *BalancerSuite) TestFailOnWrongInput() {
	pair := Pair{Base: "BAL", Quote: "USD"}

	// Wrong pair
	fr := suite.origin.Fetch([]Pair{{}})
	suite.Error(fr[0].Error)

	// Nil as a response
	fr = suite.origin.Fetch([]Pair{pair})
	suite.Equal(ErrEmptyOriginResponse, fr[0].Error)

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

	// Error during converting price to a number
	resp = &query.HTTPResponse{
		Body: []byte(`
			{
				"data": { 
					"tokenPrices":[
						{
							"poolLiquidity":"283523717.590",
							"price":"",
							"symbol":"BAL"
						}
					]
				}
			}
		`),
	}
	suite.origin.Pool.(*query.MockWorkerPool).MockResp(resp)
	fr = suite.origin.Fetch([]Pair{pair})
	suite.Error(fr[0].Error)

	// Unable to find a pair
	resp = &query.HTTPResponse{
		Body: []byte(`
			{
				"data": { 
					"tokenPrices":[
						{
							"poolLiquidity":"283523717.590",
							"price":"57.84",
							"symbol":"BTC"
						}
					]
				}
			}
		`),
	}
	suite.origin.Pool.(*query.MockWorkerPool).MockResp(resp)
	fr = suite.origin.Fetch([]Pair{pair})
	suite.Error(fr[0].Error)
}

func (suite *BalancerSuite) TestSuccessResponse() {
	pairBALUSD := Pair{Base: "BAL", Quote: "USD"}
	resp := &query.HTTPResponse{
		Body: []byte(`
			{
				"data": { 
					"tokenPrices":[
						{
							"poolLiquidity":"283523717.590",
							"price":"57.84",
							"symbol":"BAL"
						}
					]
				}
			}
		`),
	}
	suite.origin.Pool.(*query.MockWorkerPool).MockResp(resp)
	fr := suite.origin.Fetch([]Pair{pairBALUSD})

	suite.Len(fr, 1)

	// BAL/USD
	suite.NoError(fr[0].Error)
	suite.Equal(pairBALUSD, fr[0].Price.Pair)
	suite.Equal(57.84, fr[0].Price.Price)
	suite.Equal(283523717.59, fr[0].Price.Volume24h)
	suite.Greater(fr[0].Price.Timestamp.Unix(), int64(0))
}

func (suite *BalancerSuite) TestRealAPICall() {
	testRealBatchAPICall(
		suite,
		&Balancer{Pool: query.NewHTTPWorkerPool(1)},
		[]Pair{
			{Base: "BAL", Quote: "USD"},
		},
	)
}

func TestBalancerSuite(t *testing.T) {
	suite.Run(t, new(BalancerSuite))
}
