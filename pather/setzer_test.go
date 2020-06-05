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

package pather

import (
	"testing"

	"github.com/makerdao/gofer/model"
	"github.com/stretchr/testify/assert"
)

func TestPairsAndPath(t *testing.T) {
	sppf := NewSetzer()
	pairs := sppf.Pairs()
	assert.Nil(t, sppf.Path(model.NewPair("a", "z")), "Non existing pair should return nil")
	for _, p := range pairs {
		ppaths := sppf.Path(p)
		assert.NotNilf(t, ppaths, "Path should return paths for pair: %s", p)
		if ppaths != nil {
			err := model.ValidatePricePathMap(&model.PricePathMap{*p: ppaths})
			assert.NoError(t, err, "PricePaths must be valid")
		}
	}
}

func TestValidBATUSDPath(t *testing.T) {
	setzer := NewSetzer()

	paths := setzer.Path(model.NewPair("BAT", "USD"))
	assert.Greater(t, len(paths), 0)
}
