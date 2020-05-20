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

package gofer

import (
	"github.com/stretchr/testify/assert"
	"github.com/makerdao/gofer/model"
	"github.com/makerdao/gofer/query"
	"testing"
)

func TestPriceCollectorFailsWithoutWorkingPool(t *testing.T) {
	p := &model.Pair{
		Base:  "BTC",
		Quote: "ETH",
	}
	pp := &model.PotentialPricePoint{
		Exchange: &model.Exchange{
			Name: "binance",
		},
		Pair: p,
	}

	pc := &PriceCollector{}
	_, err := pc.CollectPricePoint(pp)
	assert.Error(t, err)

	// Check not started
	pc = &PriceCollector{
		wp: &query.HTTPWorkerPool{},
	}
	_, err = pc.CollectPricePoint(pp)
	assert.Error(t, err)
}
