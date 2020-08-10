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

	"github.com/makerdao/gofer"

	"github.com/stretchr/testify/assert"

	"github.com/makerdao/gofer/model"
)

func TestResponse(t *testing.T) {
	wp := NewFakeWorkerPool()
	assert.NotNil(t, wp)

	processor := gofer.NewProcessor(wp)
	assert.NotNil(t, processor)

	goferLib, err := gofer.ReadFile("./test.gofer.json")
	assert.NoError(t, err)
	assert.NotNil(t, goferLib)

	goferLib.SetProcessor(processor)

	ethUsd := model.NewPair("ETH", "USD")

	list, err := goferLib.Prices(ethUsd)
	assert.NoError(t, err)

	res, ok := list[*ethUsd]
	assert.True(t, ok)
	assert.NotNil(t, res)

	assert.Equal(t, "median", res.PriceModelName)
	assert.Equal(t, 5, len(res.Prices))

	assert.NotNil(t, res.PricePoint)
	assert.Equal(t, *ethUsd, *res.PricePoint.Pair)
	assert.Equal(t, 239.71, res.PricePoint.Price)
}
