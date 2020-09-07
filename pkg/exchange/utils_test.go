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

	"github.com/makerdao/gofer/pkg/model"
)

var testAPICalls = flag.Bool("gofer.test-api-calls", false, "enable tests on real exchanges API")

type Suite interface {
	suite.TestingSuite

	Assert() *assert.Assertions
	Exchange() Handler
}

func newPricePoint(exchangeName, base, quote string) *model.PricePoint {
	p := &model.Pair{
		Base:  base,
		Quote: quote,
	}
	return &model.PricePoint{
		Exchange: &model.Exchange{
			Name: exchangeName,
		},
		Pair: p,
	}
}

func testRealAPICall(suite Suite, exchange Handler, base, quote string) {
	if !*testAPICalls {
		suite.T().SkipNow()
	}

	suite.Assert().IsType(suite.Exchange(), exchange)

	ppp := newPricePoint("exchange", base, quote)
	pps := []*model.PricePoint{ppp}
	exchange.Fetch(pps)

	suite.Assert().NoError(pps[0].Error)
	suite.Assert().Greater(pps[0].Price, float64(0))
}
