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

	"github.com/makerdao/gofer/internal/pkg/query"
	"github.com/makerdao/gofer/pkg/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type ExchangesSuite struct {
	suite.Suite
	pool *query.MockWorkerPool
	set  *Set
}

// Setup exchange
func (suite *ExchangesSuite) SetupSuite() {
	pool := query.NewMockWorkerPool()

	suite.pool = pool
	suite.set = NewSet(map[string]Handler{
		"binance": &Binance{pool},
	})
}

func (suite *ExchangesSuite) TestCallErrorNegative() {
	res, err := suite.set.Call(nil)
	assert.Nil(suite.T(), res)
	assert.Error(suite.T(), err)

	res, err = suite.set.Call(&model.PotentialPricePoint{})
	assert.Nil(suite.T(), res)
	assert.Error(suite.T(), err)

	pp := &model.PotentialPricePoint{
		Exchange: &model.Exchange{
			Name: "unknown",
		},
	}
	res, err = suite.set.Call(pp)
	assert.Nil(suite.T(), res)
	assert.Error(suite.T(), err)
}

func (suite *ExchangesSuite) TestFailWithNilResponseForBinance() {
	pp := &model.PotentialPricePoint{
		Exchange: &model.Exchange{
			Name: "binance",
		},
		Pair: &model.Pair{
			Base:  "BTC",
			Quote: "ETH",
		},
	}

	res, err := suite.set.Call(pp)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), res)
}

func (suite *ExchangesSuite) TestSuccessBinance() {
	price := 0.024361
	json := fmt.Sprintf(`{"symbol":"ETHBTC","price":"%f"}`, price)
	resp := &query.HTTPResponse{
		Body:  []byte(json),
		Error: nil,
	}
	p := &model.Pair{
		Base:  "BTC",
		Quote: "ETH",
	}
	suite.pool.MockResp(resp)
	pp := &model.PotentialPricePoint{
		Exchange: &model.Exchange{
			Name: "binance",
		},
		Pair: p,
	}

	res, err := suite.set.Call(pp)

	assert.NoError(suite.T(), err)
	assert.EqualValues(suite.T(), p, res.Pair)
	assert.EqualValues(suite.T(), price, res.Price)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestExchangesSuite(t *testing.T) {
	suite.Run(t, new(ExchangesSuite))
}
