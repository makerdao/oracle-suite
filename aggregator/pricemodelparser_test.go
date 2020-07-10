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

package aggregator

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/makerdao/gofer/model"
)

func TestPriceModelMapParse(t *testing.T) {
	input := []byte(`{
    "b/c": {
      "method": "median",
      "sources": [
        [{"origin": "e-a", "pair": "b/c"}],
        [{"origin": "e-b", "pair": "b/c"}]
      ]
    },
    "a/c": {
      "method": "median",
      "sources": [
        [{"origin": "e-a", "pair": "a/c"}],
        [{"origin": "e-a", "pair": "a/b"}, {"origin": ".", "pair": "b/c"}]
      ]
    }
  }`)

	var pmm PriceModelMap
	err := json.Unmarshal(input, &pmm)
	assert.NoError(t, err)

	bcPriceModel := pmm[Pair{*model.NewPair("b", "c")}]
	assert.NotNil(t, bcPriceModel)
	assert.Equal(t, "median", bcPriceModel.Method)
	assert.ElementsMatch(t, []PriceRefPath{
		{{Origin: "e-a", Pair: Pair{*model.NewPair("b", "c")}}},
		{{Origin: "e-b", Pair: Pair{*model.NewPair("b", "c")}}},
	}, bcPriceModel.Sources)

	acPriceModel := pmm[Pair{*model.NewPair("a", "c")}]
	assert.NotNil(t, acPriceModel)
	assert.Equal(t, "median", acPriceModel.Method)
	assert.ElementsMatch(t, []PriceRefPath{
		{{Origin: "e-a", Pair: Pair{*model.NewPair("a", "c")}}},
		{
			{Origin: "e-a", Pair: Pair{*model.NewPair("a", "b")}},
			{Origin: ".", Pair: Pair{*model.NewPair("b", "c")}},
		},
	}, acPriceModel.Sources)
}
