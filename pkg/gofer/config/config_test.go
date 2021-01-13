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

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/makerdao/gofer/pkg/gofer/graph"
)

var Valid = Config{
	Origins: nil,
	PriceModels: map[string]PriceModel{
		"B/C": {
			Method: "median",
			Sources: [][]Source{
				{
					{Origin: "bc1", Pair: "B/C"},
				},
				{
					{Origin: "bc2", Pair: "B/C"},
				},
			},
			Params: []byte(`{"minimumSuccessfulSources": 3}`),
		},
		"A/C": {
			Method: "median",
			Sources: [][]Source{
				{
					{Origin: "ab1", Pair: "A/B"},
					{Origin: "bc1", Pair: "B/C"},
				},
				{
					{Origin: "ab2", Pair: "A/B"},
					{Origin: ".", Pair: "B/C"},
				},
			},
			Params: []byte(`{"minimumSuccessfulSources": 3}`),
		},
	},
}

var Cyclic = Config{
	Origins: nil,
	PriceModels: map[string]PriceModel{
		"A/B": {
			Method: "median",
			Sources: [][]Source{
				{
					{Origin: "ab1", Pair: "A/B"},
				},
				{
					{Origin: "ab2", Pair: "A/B"},
				},
				{
					{Origin: "ac1", Pair: "A/B"},
					{Origin: ".", Pair: "B/C"},
				},
			},
			Params: []byte(`{"minimumSuccessfulSources": 3}`),
		},
		"A/C": {
			Method: "median",
			Sources: [][]Source{
				{
					{Origin: "ab1", Pair: "A/B"},
					{Origin: ".", Pair: "B/C"},
				},
			},
			Params: []byte(`{"minimumSuccessfulSources": 3}`),
		},
		"B/C": {
			Method: "median",
			Sources: [][]Source{
				{
					{Origin: "ab1", Pair: "A/B"},
					{Origin: ".", Pair: "A/C"},
				},
			},
			Params: []byte(`{"minimumSuccessfulSources": 3}`),
		},
	},
}

func TestBuildGraphs_ValidConfig(t *testing.T) {
	j, err2 := Valid.BuildGraphs()
	assert.Nil(t, err2)

	// List of pairs used in config file:
	ab := graph.Pair{Base: "A", Quote: "B"}
	bc := graph.Pair{Base: "B", Quote: "C"}
	ac := graph.Pair{Base: "A", Quote: "C"}

	// Check if all three pairs was loaded correctly:
	assert.Contains(t, j, bc)
	assert.Contains(t, j, ac)
	assert.IsType(t, &graph.MedianAggregatorNode{}, j[bc])
	assert.IsType(t, &graph.MedianAggregatorNode{}, j[ac])

	// --- Tests for B/C pair ---
	assert.Len(t, j[bc].Children(), 2)
	// Sources have only one pair so we expect OriginNodes instead of
	// the IndirectAggregatorNode:
	assert.IsType(t, &graph.OriginNode{}, j[bc].Children()[0])
	assert.IsType(t, &graph.OriginNode{}, j[bc].Children()[1])
	// Check if pairs was assigned correctly to nodes:
	assert.Equal(t, "bc1", j[bc].Children()[0].(*graph.OriginNode).OriginPair().Origin)
	assert.Equal(t, "bc2", j[bc].Children()[1].(*graph.OriginNode).OriginPair().Origin)
	assert.Equal(t, bc, j[bc].Children()[0].(*graph.OriginNode).OriginPair().Pair)
	assert.Equal(t, bc, j[bc].Children()[1].(*graph.OriginNode).OriginPair().Pair)

	// --- Tests for A/C pair ---
	assert.Len(t, j[ac].Children(), 2)
	// Sources have more than one pair so now we expect the
	// IndirectAggregatorNode.
	assert.IsType(t, &graph.IndirectAggregatorNode{}, j[ac].Children()[0])
	assert.IsType(t, &graph.IndirectAggregatorNode{}, j[ac].Children()[1])
	// Check if pairs was assigned correctly to nodes:
	assert.Equal(t, ac, j[ac].Children()[0].(*graph.IndirectAggregatorNode).Pair())
	assert.Equal(t, ac, j[ac].Children()[1].(*graph.IndirectAggregatorNode).Pair())
	assert.Equal(t, ab, j[ac].Children()[0].(*graph.IndirectAggregatorNode).Children()[0].(*graph.OriginNode).OriginPair().Pair)
	assert.Equal(t, bc, j[ac].Children()[0].(*graph.IndirectAggregatorNode).Children()[1].(*graph.OriginNode).OriginPair().Pair)
	// In a second source, there is a reference to another root node. We should
	// use previously created instance instead creating new one:
	assert.Same(t, j[bc], j[ac].Children()[1].(*graph.IndirectAggregatorNode).Children()[1])
}

func TestBuildGraphs_CyclicConfig(t *testing.T) {
	_, err2 := Cyclic.BuildGraphs()
	assert.Error(t, err2)
}
