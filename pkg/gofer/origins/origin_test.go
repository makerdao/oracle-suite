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

	"github.com/makerdao/oracle-suite/internal/query"
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
	cr := suite.set.Fetch(map[string][]Pair{"x": {{}}})
	assert.Error(suite.T(), cr["x"][0].Error)

	pair := Pair{Quote: "A", Base: "B"}
	cr = suite.set.Fetch(map[string][]Pair{"x": {pair}})

	assert.Equal(suite.T(), pair, cr["x"][0].Price.Pair)
	assert.Error(suite.T(), cr["x"][0].Error)
}

func (suite *OriginsSuite) TestFailWithNilResponseForBinance() {
	resp := &query.HTTPResponse{
		Body:  []byte{},
		Error: nil,
	}

	suite.pool.MockResp(resp)

	pair := Pair{Base: "BTC", Quote: "ETH"}
	cr := suite.set.Fetch(map[string][]Pair{"binance": {pair}})

	assert.Error(suite.T(), cr["binance"][0].Error)
}

func (suite *OriginsSuite) TestSuccessBinance() {
	price := 0.024361
	json := fmt.Sprintf(`[{"symbol":"ETHBTC","lastPrice":"%f"}]`, price)
	resp := &query.HTTPResponse{
		Body:  []byte(json),
		Error: nil,
	}

	suite.pool.MockResp(resp)

	pair := Pair{Quote: "BTC", Base: "ETH"}
	cr := suite.set.Fetch(map[string][]Pair{"binance": {pair}})

	assert.NoError(suite.T(), cr["binance"][0].Error)
	assert.EqualValues(suite.T(), pair, cr["binance"][0].Price.Pair)
	assert.EqualValues(suite.T(), price, cr["binance"][0].Price.Price)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestOriginsSuite(t *testing.T) {
	suite.Run(t, new(OriginsSuite))
}

func TestGettingContractAddressByBaseAndQuote(t *testing.T) {
	contract := "0x0000"
	contracts := ContractAddresses{"BTC/ETH": contract}

	// Not existing pair
	address, ok := contracts.ByPair(Pair{Base: "BTC", Quote: "USD"})
	assert.Empty(t, address)
	assert.False(t, ok)

	// Existing direct pair
	address, ok = contracts.ByPair(Pair{Base: "BTC", Quote: "ETH"})
	assert.True(t, ok)
	assert.Equal(t, contract, address)

	// Existing reversed pair
	address, ok = contracts.ByPair(Pair{Base: "ETH", Quote: "BTC"})
	assert.True(t, ok)
	assert.Equal(t, contract, address)
}

func TestReplacingSymbolUsingAliases(t *testing.T) {
	aliases := SymbolAliases{"ETH": "WETH"}

	// Should not be changed
	symbol := aliases.Replace("BTC")
	assert.Equal(t, "BTC", symbol)

	// Should be replaced
	symbol = aliases.Replace("ETH")
	assert.Equal(t, "WETH", symbol)

	pair := Pair{Base: "BTC", Quote: "ETH"}
	replaced := aliases.ReplacePair(pair)
	assert.Equal(t, "BTC", replaced.Base)
	assert.Equal(t, "WETH", replaced.Quote)
}
