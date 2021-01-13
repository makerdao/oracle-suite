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

	"github.com/makerdao/gofer/pkg/gofer/graph"

	"github.com/makerdao/gofer/internal/gofer/marshal/testutil"
)

func TestTrace_Graph(t *testing.T) {
	disableColors()

	g := testutil.Graph(graph.Pair{Base: "A", Quote: "B"})
	j := newTrace()

	err := j.Write(g)
	assert.NoError(t, err)

	err = j.Close()
	assert.NoError(t, err)

	b, err := ioutil.ReadAll(j)
	assert.NoError(t, err)

	expected := `
Graph for A/B:
───graph.MedianAggregatorNode(pair:A/B)
   ├──graph.OriginNode(pair:A/B, origin:a)
   ├──graph.IndirectAggregatorNode(pair:A/B)
   │  └──graph.OriginNode(pair:A/B, origin:a)
   └──graph.MedianAggregatorNode(pair:A/B)
      ├──graph.OriginNode(pair:A/B, origin:a)
      └──graph.OriginNode(pair:A/B, origin:b)
`

	assert.Equal(t, expected[1:], string(b))
}

func TestTrace_Ticks(t *testing.T) {
	disableColors()

	g := testutil.Graph(graph.Pair{Base: "A", Quote: "B"})
	j := newTrace()

	err := j.Write(g.Tick())
	assert.NoError(t, err)

	err = j.Close()
	assert.NoError(t, err)

	b, err := ioutil.ReadAll(j)
	assert.NoError(t, err)

	expected := `
Price for A/B:
───AggregatorTick(pair:A/B, price:10, timestamp:1970-01-01T00:00:10Z, method:median, min:1)
   ├──OriginTick(pair:A/B, origin:a, price:10, timestamp:1970-01-01T00:00:10Z)
   ├──AggregatorTick(pair:A/B, price:10, timestamp:1970-01-01T00:00:10Z, method:indirect)
   │  └──OriginTick(pair:A/B, origin:a, price:10, timestamp:1970-01-01T00:00:10Z)
   └──AggregatorTick(pair:A/B, price:10, timestamp:1970-01-01T00:00:10Z, method:median, min:1)
      ├──OriginTick(pair:A/B, origin:a, price:10, timestamp:1970-01-01T00:00:10Z)
      └──[ERROR] OriginTick(pair:A/B, origin:b, price:20, timestamp:1970-01-01T00:00:20Z)
            Error: something
`

	assert.Equal(t, expected[1:], string(b))
}

func TestTrace_Origins(t *testing.T) {
	disableColors()

	p := graph.Pair{Base: "A", Quote: "B"}
	j := newTrace()

	err := j.Write(map[graph.Pair][]string{
		p: {"a", "b", "c"},
	})

	assert.NoError(t, err)

	err = j.Close()
	assert.NoError(t, err)

	b, err := ioutil.ReadAll(j)
	assert.NoError(t, err)

	expected := `
───A/B
   ├──a
   ├──b
   └──c
`

	assert.Equal(t, expected[1:], string(b))
}
