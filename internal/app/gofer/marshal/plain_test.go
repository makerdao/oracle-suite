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

package marshal

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/makerdao/gofer/pkg/aggregator"
	"github.com/makerdao/gofer/pkg/model"
)

func TestPlainForPriceAggregate(t *testing.T) {
	pa := model.NewPriceAggregate(
		"agg",
		&model.PricePoint{
			Price: 1.0,
			Pair:  &model.Pair{Base: "a", Quote: "b"},
		},
		model.NewPriceAggregate(
			"med",
			&model.PricePoint{
				Price: 2.0,
				Pair:  &model.Pair{Base: "c", Quote: "d"},
			},
		),
	)

	j := newPlain()
	assert.Nil(t, j.Write(pa, nil))
	assert.Nil(t, j.Write(pa, nil))
	assert.Nil(t, j.Close())

	r, err := ioutil.ReadAll(j)
	assert.Nil(t, err)

	assert.Equal(t, "a/b 1.000000\na/b 1.000000\n", string(r))
}

func TestPlainForExchange(t *testing.T) {
	ex := &model.Exchange{
		Name:   "foobar",
		Config: map[string]string{"foo": "bar"},
	}

	j := newPlain()
	assert.Nil(t, j.Write(ex, nil))
	assert.Nil(t, j.Write(ex, nil))
	assert.Nil(t, j.Close())

	r, err := ioutil.ReadAll(j)
	assert.Nil(t, err)

	assert.Equal(t, "foobar\nfoobar\n", string(r))
}

func TestPlainForPriceModelMap(t *testing.T) {
	pmm := aggregator.PriceModelMap{
		aggregator.Pair{Pair: model.Pair{Base: "A", Quote: "B"}}: aggregator.PriceModel{
			Method:     "method",
			MinSources: 2,
			Sources: []aggregator.PriceRefPath{
				[]aggregator.PriceRef{{
					Origin: "origin",
					Pair:   aggregator.Pair{Pair: model.Pair{Base: "A", Quote: "B"}},
				}},
			},
		},
	}

	j := newPlain()
	assert.Nil(t, j.Write(pmm, nil))
	assert.Nil(t, j.Write(pmm, nil))
	assert.Nil(t, j.Close())

	r, err := ioutil.ReadAll(j)
	assert.Nil(t, err)

	assert.Equal(t, "A/B\nA/B\n", string(r))
}
