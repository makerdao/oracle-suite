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

type CurveSuite struct {
	suite.Suite
	origin *BaseExchangeHandler
}

func (suite *CurveSuite) Origin() Handler {
	return suite.origin
}

func (suite *CurveSuite) SetupSuite() {
	suite.origin = NewBaseExchangeHandler(
		Curve{WorkerPool: query.NewMockWorkerPool()},
		nil,
	)
}

func (suite *CurveSuite) TestFailOnWrongInput() {
	pair := Pair{Base: "stETH", Quote: "ETH"}

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

	suite.origin.Pool().(*query.MockWorkerPool).MockResp(resp)
	fr = suite.origin.Fetch([]Pair{pair})
	suite.Equal(ourErr, fr[0].Error)

	// Unable to find a pair
	resp = &query.HTTPResponse{
		Body: []byte(`
			{
				"data": {
					"pool": null
				}
			}
		`),
	}
	suite.origin.Pool().(*query.MockWorkerPool).MockResp(resp)
	fr = suite.origin.Fetch([]Pair{pair})
	suite.Error(fr[0].Error)
}

func (suite *CurveSuite) TestSuccessResponse() {
	pairBALUSD := Pair{Base: "stETH", Quote: "ETH"}
	resp := &query.HTTPResponse{
		Body: []byte(`
			{
				"data": {
					"pool": {
						"id": "0xdc24316b9ae028f1497c275eb9192a3ea0f67022",
						"A": "5000",
						"coins": [
							{
								"balance": "576085.752212471908264349",
								"rate": "1",
								"token": {
									"symbol": "ETH"
								}
							},
							{
								"balance": "721671.125949688906012467",
								"rate": "1",
								"token": {
									"symbol": "stETH"
								}
							}
						],
						"hourlyVolumes": [
							{"volume": "1.0"},
							{"volume": "1.0"},
							{"volume": "1.0"},
							{"volume": "1.0"},
							{"volume": "1.0"},
							{"volume": "1.0"},
							{"volume": "1.0"},
							{"volume": "1.0"},
							{"volume": "1.0"},
							{"volume": "1.0"},
							{"volume": "1.0"},
							{"volume": "1.0"},
							{"volume": "1.0"},
							{"volume": "1.0"},
							{"volume": "1.0"},
							{"volume": "1.0"},
							{"volume": "1.0"},
							{"volume": "1.0"},
							{"volume": "1.0"},
							{"volume": "1.0"},
							{"volume": "1.0"},
							{"volume": "1.0"},
							{"volume": "1.0"},
							{"volume": "1.0"}
						]
					}
				}
			}
		`),
	}
	suite.origin.Pool().(*query.MockWorkerPool).MockResp(resp)
	fr := suite.origin.Fetch([]Pair{pairBALUSD})

	suite.Len(fr, 1)

	// BAL/USD
	suite.NoError(fr[0].Error)
	suite.Equal(pairBALUSD, fr[0].Price.Pair)
	suite.InDelta(0.9955018691252917, fr[0].Price.Price, 0.00001)
	suite.Equal(float64(24), fr[0].Price.Volume24h)
	suite.Greater(fr[0].Price.Timestamp.Unix(), int64(0))
}

func (suite *CurveSuite) TestRealAPICall() {
	origin := NewBaseExchangeHandler(
		Curve{WorkerPool: query.NewHTTPWorkerPool(1)},
		nil,
	)

	testRealBatchAPICall(
		suite,
		origin,
		[]Pair{
			{Base: "stETH", Quote: "ETH"},
		},
	)
}

func TestCurveSuite(t *testing.T) {
	suite.Run(t, new(CurveSuite))
}
