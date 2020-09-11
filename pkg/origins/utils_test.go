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
	"flag"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

var testAPICalls = flag.Bool("gofer.test-api-calls", false, "enable tests on real exchanges API")

type Suite interface {
	suite.TestingSuite

	Assert() *assert.Assertions
	Exchange() Handler
}

func testRealAPICall(suite Suite, exchange Handler, base, quote string) {
	if !*testAPICalls {
		suite.T().SkipNow()
	}

	suite.Assert().IsType(suite.Exchange(), exchange)

	pair := Pair{Base: base, Quote: quote}
	cr := exchange.Fetch([]Pair{pair})

	suite.Assert().NoError(cr[0].Error)
	suite.Assert().Greater(cr[0].Tick.Price, float64(0))
}
