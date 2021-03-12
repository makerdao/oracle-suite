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

type UniswapSuite struct {
	suite.Suite
	origin *Uniswap
}

func (suite *UniswapSuite) Origin() Handler {
	return suite.origin
}

func (suite *UniswapSuite) SetupSuite() {
	suite.origin = &Uniswap{Pool: query.NewMockWorkerPool()}
}

func (suite *UniswapSuite) TestFailOnWrongInput() {
	pair := Pair{Base: "LRC", Quote: "WETH"}

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
					"pairs": [
						{
							"id": "0x8878df9e1a7c87dcbf6d3999d997f262c05d8c70",
							"token0Price": "",
							"token1Price": "",
							"volumeToken0": "274940368.686748844780986508",
							"volumeToken1": "142365.832159709562349781",
							"token0": {
								"symbol": "LRC"
							},
							"token1": {
								"symbol": "WETH"
							}
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
					"pairs": [
						{
							"id": "0xaabbccddeeffgghh0011223344556677889900aa",
							"token0Price": "1560.208506522844994633814164798516",
							"token1Price": "0.0006409399742529590737926118088434103",
							"volumeToken0": "274940368.686748844780986508",
							"volumeToken1": "142365.832159709562349781",
							"token0": {
								"symbol": "LRC"
							},
							"token1": {
								"symbol": "WBTC"
							}
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

func (suite *UniswapSuite) TestSuccessResponse() {
	pairLRCWETH := Pair{Base: "LRC", Quote: "WETH"}
	pairWETHCOMP := Pair{Base: "WETH", Quote: "COMP"}

	resp := &query.HTTPResponse{
		Body: []byte(`
			{
				"data": {
					"pairs": [
						{
							"id": "0x8878df9e1a7c87dcbf6d3999d997f262c05d8c70",
							"token0Price": "1560.2121",
							"token1Price": "0.0006",
							"volumeToken0": "274940368.6801",
							"volumeToken1": "142365.8321",
							"token0": {
								"symbol": "LRC"
							},
							"token1": {
								"symbol": "WETH"
							}
						},
						{
							"id": "0xcffdded873554f362ac02f8fb1f02e5ada10516f",
							"token0Price": "2.4889",
							"token1Price": "0.4017",
							"volumeToken0": "1295833.9715",
							"volumeToken1": "714460.7483",
							"token0": {
								"symbol": "COMP"
							},
							"token1": {
								"symbol": "WETH"
							}
						},
						{
							"id": "0xf49c43ae0faf37217bdcb00df478cf793edd6687",
							"token0Price": "1560.2085",
							"token1Price": "0.0006",
							"volumeToken0": "274940368.6867",
							"volumeToken1": "142365.8321",
							"token0": {
								"symbol": "KNC"
							},
							"token1": {
								"symbol": "WETH"
							}
						}
					]
				}
			}
		`),
	}
	suite.origin.Pool.(*query.MockWorkerPool).MockResp(resp)
	fr := suite.origin.Fetch([]Pair{pairLRCWETH, pairWETHCOMP})

	suite.Len(fr, 2)

	// LRC/WETH
	suite.NoError(fr[0].Error)
	suite.Equal(pairLRCWETH, fr[0].Price.Pair)
	suite.Equal(0.0006, fr[0].Price.Price)
	suite.Equal(0.0006, fr[0].Price.Bid)
	suite.Equal(0.0006, fr[0].Price.Ask)
	suite.Equal(274940368.6801, fr[0].Price.Volume24h)
	suite.Greater(fr[0].Price.Timestamp.Unix(), int64(0))

	// WETH/COMP
	suite.NoError(fr[1].Error)
	suite.Equal(pairWETHCOMP, fr[1].Price.Pair)
	suite.Equal(2.4889, fr[1].Price.Price)
	suite.Equal(2.4889, fr[1].Price.Bid)
	suite.Equal(2.4889, fr[1].Price.Ask)
	suite.Equal(714460.7483, fr[1].Price.Volume24h)
	suite.Greater(fr[1].Price.Timestamp.Unix(), int64(0))
}

func (suite *UniswapSuite) TestRealAPICall() {
	testRealBatchAPICall(
		suite,
		&Uniswap{Pool: query.NewHTTPWorkerPool(1)},
		[]Pair{
			{Base: "LRC", Quote: "WETH"},
			{Base: "WETH", Quote: "KNC"},
		},
	)
}

func TestUniswapSuite(t *testing.T) {
	suite.Run(t, new(UniswapSuite))
}
