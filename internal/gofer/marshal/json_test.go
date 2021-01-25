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
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/makerdao/gofer/pkg/gofer/graph"

	"github.com/makerdao/gofer/internal/gofer/marshal/testutil"
)

func TestJSON_Graph(t *testing.T) {
	var err error
	j := newJSON(false)

	err = j.Write(testutil.Graph(graph.Pair{Base: "A", Quote: "B"}))
	assert.NoError(t, err)

	err = j.Write(testutil.Graph(graph.Pair{Base: "C", Quote: "D"}))
	assert.NoError(t, err)

	err = j.Close()
	assert.NoError(t, err)

	b, err := ioutil.ReadAll(j)
	assert.NoError(t, err)

	expected := `["A/B", "C/D"]`

	assert.JSONEq(t, expected, string(b))
}

func TestNDJSON_Graph(t *testing.T) {
	var err error
	j := newJSON(true)

	err = j.Write(testutil.Graph(graph.Pair{Base: "A", Quote: "B"}))
	assert.NoError(t, err)

	err = j.Write(testutil.Graph(graph.Pair{Base: "C", Quote: "D"}))
	assert.NoError(t, err)

	err = j.Close()
	assert.NoError(t, err)

	b, err := ioutil.ReadAll(j)
	assert.NoError(t, err)

	result := bytes.Split(b, []byte("\n"))

	assert.JSONEq(t, `"A/B"`, string(result[0]))
	assert.JSONEq(t, `"C/D"`, string(result[1]))
}

func TestJSON_Ticks(t *testing.T) {
	g := testutil.Graph(graph.Pair{Base: "A", Quote: "B"})
	j := newJSON(false)

	err := j.Write(g.Tick())
	assert.NoError(t, err)

	err = j.Close()
	assert.NoError(t, err)

	b, err := ioutil.ReadAll(j)
	assert.NoError(t, err)

	expected := `
		[
		   {
			  "type":"aggregate",
			  "params":{
				 "method":"median",
				 "min":"1"
			  },
			  "base":"A",
			  "quote":"B",
			  "price":10,
			  "bid":10,
			  "ask":10,
			  "ts":"1970-01-01T00:00:10Z",
			  "ticks":[
				 {
					"type":"origin",
					"origin":"a",
					"base":"A",
					"quote":"B",
					"price":10,
					"bid":10,
					"ask":10,
					"vol24h":10,
					"ts":"1970-01-01T00:00:10Z"
				 },
				 {
					"type":"aggregate",
					"params":{
					   "method":"indirect"
					},
					"base":"A",
					"quote":"B",
					"price":10,
					"bid":10,
					"ask":10,
					"vol24h":10,
					"ts":"1970-01-01T00:00:10Z",
					"ticks":[
					   {
						  "type":"origin",
						  "origin":"a",
						  "base":"A",
						  "quote":"B",
						  "price":10,
						  "bid":10,
						  "ask":10,
						  "vol24h":10,
						  "ts":"1970-01-01T00:00:10Z"
					   }
					]
				 },
				 {
					"type":"aggregate",
					"params":{
					   "method":"median",
					   "min":"1"
					},
					"base":"A",
					"quote":"B",
					"price":10,
					"bid":10,
					"ask":10,
					"ts":"1970-01-01T00:00:10Z",
					"ticks":[
					   {
						  "type":"origin",
						  "origin":"a",
						  "base":"A",
						  "quote":"B",
						  "price":10,
						  "bid":10,
						  "ask":10,
						  "vol24h":10,
						  "ts":"1970-01-01T00:00:10Z"
					   },
					   {
						  "type":"origin",
						  "origin":"b",
						  "base":"A",
						  "quote":"B",
						  "price":20,
						  "bid":20,
						  "ask":20,
						  "vol24h":20,
						  "ts":"1970-01-01T00:00:20Z",
						  "error": "something"
					   }
					]
				 }
			  ]
		   }
		]
	`

	assert.JSONEq(t, expected, string(b))
}

func TestJSON_Origins(t *testing.T) {
	var err error
	j := newJSON(false)

	ab := graph.Pair{Base: "A", Quote: "B"}
	cd := graph.Pair{Base: "C", Quote: "D"}

	err = j.Write(map[graph.Pair][]string{
		ab: {"a", "b", "c"},
		cd: {"x", "y", "z"},
	})

	assert.NoError(t, err)

	err = j.Close()
	assert.NoError(t, err)

	b, err := ioutil.ReadAll(j)
	assert.NoError(t, err)

	expected := `[{"A/B":["a","b","c"], "C/D":["x","y","z"]}]`

	assert.JSONEq(t, expected, string(b))
}

func TestNDJSON_Origins(t *testing.T) {
	var err error
	j := newJSON(true)

	ab := graph.Pair{Base: "A", Quote: "B"}
	cd := graph.Pair{Base: "C", Quote: "D"}

	err = j.Write(map[graph.Pair][]string{
		ab: {"a", "b", "c"},
	})
	assert.NoError(t, err)

	err = j.Write(map[graph.Pair][]string{
		cd: {"x", "y", "z"},
	})
	assert.NoError(t, err)

	err = j.Close()
	assert.NoError(t, err)

	b, err := ioutil.ReadAll(j)
	assert.NoError(t, err)

	result := bytes.Split(b, []byte("\n"))

	assert.JSONEq(t, `{"A/B":["a","b","c"]}`, string(result[0]))
	assert.JSONEq(t, `{"C/D":["x","y","z"]}`, string(result[1]))
}
