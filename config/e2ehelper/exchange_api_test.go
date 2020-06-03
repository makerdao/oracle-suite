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

package e2ehelper

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/makerdao/gofer/exchange"
	"github.com/makerdao/gofer/model"
	"github.com/makerdao/gofer/query"
)

func TestResponse(t *testing.T) {
	wp := NewFakeWorkerPool()

	assert.NotNil(t, wp)

	ethUsd := &model.PotentialPricePoint{
		Pair: &model.Pair{
			Base:  "ETH",
			Quote: "USD",
		},
	}
	bitstamp := &exchange.Bitstamp{}
	resp := wp.Query(&query.HTTPRequest{URL: bitstamp.GetURL(ethUsd)})

	assert.NotNil(t, resp)
	t.Log(string(resp.Body))
}
