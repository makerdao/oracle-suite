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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/makerdao/gofer/internal/query"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type OriginsSuite struct {
	suite.Suite
	pool *query.MockWorkerPool
	set  *Set
}

// Setup origin
func (suite *OriginsSuite) SetupSuite() {
	pool := query.NewMockWorkerPool()

	suite.pool = pool
	suite.set = NewSet(map[string]Handler{
		"binance": &Binance{pool},
	})
}

func (suite *OriginsSuite) TestCallWithMissingOrigin() {
	cr := suite.set.Call(map[string][]Pair{"x": {{}}})
	assert.Error(suite.T(), cr["x"][0].Error)

	pair := Pair{Quote: "A", Base: "B"}
	cr = suite.set.Call(map[string][]Pair{"x": {pair}})

	assert.Equal(suite.T(), pair, cr["x"][0].Tick.Pair)
	assert.Error(suite.T(), cr["x"][0].Error)
}

func (suite *OriginsSuite) TestFailWithNilResponseForBinance() {
	resp := &query.HTTPResponse{
		Body:  []byte{},
		Error: nil,
	}

	suite.pool.MockResp(resp)

	pair := Pair{Base: "BTC", Quote: "ETH"}
	cr := suite.set.Call(map[string][]Pair{"binance": {pair}})

	assert.Error(suite.T(), cr["binance"][0].Error)
}

func (suite *OriginsSuite) TestSuccessBinance() {
	price := 0.024361
	json := fmt.Sprintf(`{"symbol":"ETHBTC","price":"%f"}`, price)
	resp := &query.HTTPResponse{
		Body:  []byte(json),
		Error: nil,
	}

	suite.pool.MockResp(resp)

	pair := Pair{Quote: "BTC", Base: "ETH"}
	cr := suite.set.Call(map[string][]Pair{"binance": {pair}})

	assert.NoError(suite.T(), cr["binance"][0].Error)
	assert.EqualValues(suite.T(), pair, cr["binance"][0].Tick.Pair)
	assert.EqualValues(suite.T(), price, cr["binance"][0].Tick.Price)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestOriginsSuite(t *testing.T) {
	suite.Run(t, new(OriginsSuite))
}
