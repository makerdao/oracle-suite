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

	"github.com/makerdao/gofer/internal/query"
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
	pps := []*model.PricePoint{{}}
	suite.set.Fetch(pps)
	assert.Error(suite.T(), pps[0].Error)

	ex := &model.Exchange{Name: "unknown"}
	pp := &model.PricePoint{
		Exchange: ex,
	}
	pps = []*model.PricePoint{pp}
	suite.set.Fetch(pps)
	assert.Same(suite.T(), ex, pps[0].Exchange)
	assert.Error(suite.T(), pps[0].Error)
}

func (suite *ExchangesSuite) TestFailWithNilResponseForBinance() {
	pp := &model.PricePoint{
		Exchange: &model.Exchange{
			Name: "binance",
		},
		Pair: &model.Pair{
			Base:  "BTC",
			Quote: "ETH",
		},
	}

	pps := []*model.PricePoint{pp}
	suite.set.Fetch(pps)

	assert.Error(suite.T(), pps[0].Error)
	assert.NotNil(suite.T(), pps[0])
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
	pp := &model.PricePoint{
		Exchange: &model.Exchange{
			Name: "binance",
		},
		Pair: p,
	}

	pps := []*model.PricePoint{pp}
	suite.set.Fetch(pps)

	assert.NoError(suite.T(), pps[0].Error)
	assert.EqualValues(suite.T(), p, pps[0].Pair)
	assert.EqualValues(suite.T(), price, pps[0].Price)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestExchangesSuite(t *testing.T) {
	suite.Run(t, new(ExchangesSuite))
}
