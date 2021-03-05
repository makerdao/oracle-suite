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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/makerdao/gofer/internal/gofer/marshal/testutil"
	"github.com/makerdao/gofer/pkg/gofer"
)

func TestJSON_Nodes(t *testing.T) {
	var err error
	m := newJSON(false)

	ab := gofer.Pair{Base: "A", Quote: "B"}
	cd := gofer.Pair{Base: "C", Quote: "D"}
	ns := testutil.Nodes(ab, cd)

	err = m.Write(ns[ab])
	assert.NoError(t, err)

	err = m.Write(ns[cd])
	assert.NoError(t, err)

	b, err := m.Bytes()
	assert.NoError(t, err)

	expected := `["A/B", "C/D"]`

	assert.JSONEq(t, expected, string(b))
}

func TestNDJSON_Nodes(t *testing.T) {
	var err error
	m := newJSON(true)

	ab := gofer.Pair{Base: "A", Quote: "B"}
	cd := gofer.Pair{Base: "C", Quote: "D"}
	ns := testutil.Nodes(ab, cd)

	err = m.Write(ns[ab])
	assert.NoError(t, err)

	err = m.Write(ns[cd])
	assert.NoError(t, err)

	b, err := m.Bytes()
	assert.NoError(t, err)

	result := bytes.Split(b, []byte("\n"))

	assert.JSONEq(t, `"A/B"`, string(result[0]))
	assert.JSONEq(t, `"C/D"`, string(result[1]))
}

func TestJSON_Ticks(t *testing.T) {
	var err error
	m := newJSON(false)

	ab := gofer.Pair{Base: "A", Quote: "B"}
	ts := testutil.Ticks(ab)

	err = m.Write(ts[ab])
	assert.NoError(t, err)

	b, err := m.Bytes()
	assert.NoError(t, err)

	expected := `
		[
		   {
			  "type":"aggregator",
			  "base":"A",
			  "quote":"B",
			  "price":10,
			  "bid":10,
			  "ask":10,
			  "vol24h":0,
			  "ts":"1970-01-01T00:00:10Z",
			  "params":{
				 "method":"median",
				 "min":"1"
			  },
			  "ticks":[
				 {
					"type":"origin",
					"base":"A",
					"quote":"B",
					"price":10,
					"bid":10,
					"ask":10,
					"vol24h":10,
					"ts":"1970-01-01T00:00:10Z",
					"params":{
					   "origin":"a"
					}
				 },
				 {
					"type":"aggregator",
					"base":"A",
					"quote":"B",
					"price":10,
					"bid":10,
					"ask":10,
					"vol24h":10,
					"ts":"1970-01-01T00:00:10Z",
					"params":{
					   "method":"indirect"
					},
					"ticks":[
					   {
						  "type":"origin",
						  "base":"A",
						  "quote":"B",
						  "price":10,
						  "bid":10,
						  "ask":10,
						  "vol24h":10,
						  "ts":"1970-01-01T00:00:10Z",
						  "params":{
							 "origin":"a"
						  }
					   }
					]
				 },
				 {
					"type":"aggregator",
					"base":"A",
					"quote":"B",
					"price":10,
					"bid":10,
					"ask":10,
					"vol24h":0,
					"ts":"1970-01-01T00:00:10Z",
					"params":{
					   "method":"median",
					   "min":"1"
					},
					"ticks":[
					   {
						  "type":"origin",
						  "base":"A",
						  "quote":"B",
						  "price":10,
						  "bid":10,
						  "ask":10,
						  "vol24h":10,
						  "ts":"1970-01-01T00:00:10Z",
						  "params":{
							 "origin":"a"
						  }
					   },
					   {
						  "type":"origin",
						  "base":"A",
						  "quote":"B",
						  "price":20,
						  "bid":20,
						  "ask":20,
						  "vol24h":20,
						  "ts":"1970-01-01T00:00:20Z",
						  "params":{
							 "origin":"b"
						  },
						  "error":"something"
					   }
					]
				 }
			  ]
		   }
		]
	`

	assert.JSONEq(t, expected, string(b))
}
