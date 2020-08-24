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
	"flag"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/makerdao/gofer/model"
	"github.com/makerdao/gofer/query"
)

var testAPICalls = flag.Bool("gofer.test-api-calls", false, "enable tests on real exchanges API")

type Suite interface {
	suite.TestingSuite

	Assert() *assert.Assertions
	Exchange() Handler
}

func newPotentialPricePoint(exchangeName, base, quote string) *model.PotentialPricePoint {
	p := &model.Pair{
		Base:  base,
		Quote: quote,
	}
	return &model.PotentialPricePoint{
		Exchange: &model.Exchange{
			Name: exchangeName,
		},
		Pair: p,
	}
}

func testRealAPICall(suite Suite, base, quote string) {
	if !*testAPICalls {
		suite.T().SkipNow()
	}

	wp := query.NewHTTPWorkerPool(1)
	wp.Start()
	ppp := newPotentialPricePoint("exchange", base, quote)
	pp, err := suite.Exchange().Call(ppp)
	suite.Assert().NoError(wp.Stop())

	suite.Assert().NoError(err)
	suite.Assert().Greater(pp.Price, float64(0))
}
