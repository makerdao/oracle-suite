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
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/makerdao/gofer/pkg/graph"
)

func TestParseJSONFile_ValidConfig(t *testing.T) {
	_, err := ParseJSONFile("./testdata/config.valid.json")
	assert.NoError(t, err)
}

func TestParseJSONFile_MissingFile(t *testing.T) {
	_, err := ParseJSONFile("missing")
	assert.Error(t, err)
}

func TestBuildGraphs_ValidConfig(t *testing.T) {
	f, err1 := ParseJSONFile("./testdata/config.valid.json")
	assert.Nil(t, err1)

	j, err2 := f.BuildGraphs()
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
	f, err1 := ParseJSONFile("./testdata/config.cyclic.json")
	assert.Nil(t, err1)

	_, err2 := f.BuildGraphs()
	assert.Error(t, err2)
}

func TestBuildGraphs_NoSources(t *testing.T) {
	// Price model without sources should be parsed without an error:
	f, err1 := ParseJSON([]byte(`
		{
		  "pricemodels": {
			"A/B": {
			  "method": "median",
			  "sources": []
			}
		  }
		}
	`))
	assert.Nil(t, err1)

	_, err2 := f.BuildGraphs()
	assert.Nil(t, err2)
}

func TestBuildGraphs_MissingParams(t *testing.T) {
	// The "params" may be skipped:
	f, err1 := ParseJSON([]byte(`
		{
		  "pricemodels": {
			"A/B": {
			  "method": "median",
			  "sources": [
				[{"origin": "ab", "pair": "A/B"}]
			  ]
			}
		  }
		}
	`))
	assert.Nil(t, err1)

	_, err2 := f.BuildGraphs()
	assert.Nil(t, err2)
}

func TestBuildGraphs_InvalidPairName(t *testing.T) {
	// The "A_B" is incorrect, "A/B" should be used instead:
	f, err1 := ParseJSON([]byte(`
		{
		  "pricemodels": {
			"A_B": {
			  "method": "median",
			  "sources": []
			}
		  }
		}
	`))
	assert.Nil(t, err1)

	_, err2 := f.BuildGraphs()
	assert.Error(t, err2)
}

func TestBuildGraphs_ReferenceToMissingPair(t *testing.T) {
	f, err1 := ParseJSON([]byte(`
		{
		  "pricemodels": {
			"A/B": {
			  "method": "median",
			  "sources": [
				[{"origin": ".", "pair": "X/Y"}]
			  ]
			}
		  }
		}
	`))
	assert.Nil(t, err1)

	_, err2 := f.BuildGraphs()
	assert.Error(t, err2)
}

func TestBuildGraphs_ReferenceToSelf(t *testing.T) {
	f, err1 := ParseJSON([]byte(`
		{
		  "pricemodels": {
			"A/B": {
			  "method": "median",
			  "sources": [
				[{"origin": ".", "pair": "A/B"}]
			  ]
			}
		  }
		}
	`))
	assert.Nil(t, err1)

	_, err2 := f.BuildGraphs()
	assert.Error(t, err2)
}

func TestBuildGraphs_DefaultTTL(t *testing.T) {
	f, err1 := ParseJSON([]byte(`
		{
		  "pricemodels": {
			"A/B": {
			  "method": "median",
			  "sources": [
				[{"origin": "ab", "pair": "A/B"}]
			  ]
			}
		  }
		}
	`))
	assert.Nil(t, err1)

	p, _ := graph.NewPair("A/B")
	g, _ := f.BuildGraphs()

	assert.Equal(t, defaultMaxTTL, g[p].Children()[0].(*graph.OriginNode).MaxTTL())
	assert.Equal(t, defaultMaxTTL - minTTLDifference, g[p].Children()[0].(*graph.OriginNode).MinTTL())
}

func TestBuildGraphs_OriginTTL(t *testing.T) {
	f, err1 := ParseJSON([]byte(`
		{
		  "pricemodels": {
			"A/B": {
			  "method": "median",
			  "sources": [
				[{"origin": "ab", "pair": "A/B", "ttl": 120}]
			  ]
			}
		  }
		}
	`))
	assert.Nil(t, err1)

	p, _ := graph.NewPair("A/B")
	g, _ := f.BuildGraphs()

	assert.Equal(t, 120 * time.Second, g[p].Children()[0].(*graph.OriginNode).MaxTTL())
	assert.Equal(t, 90 * time.Second, g[p].Children()[0].(*graph.OriginNode).MinTTL())
}

func TestBuildGraphs_MedianTTL(t *testing.T) {
	f, err1 := ParseJSON([]byte(`
		{
		  "pricemodels": {
			"A/B": {
			  "method": "median",
			  "ttl": 120,
			  "sources": [
				[{"origin": "ab", "pair": "A/B"}]
			  ]
			}
		  }
		}
	`))
	assert.Nil(t, err1)

	p, _ := graph.NewPair("A/B")
	g, _ := f.BuildGraphs()

	assert.Equal(t, 120 * time.Second, g[p].Children()[0].(*graph.OriginNode).MaxTTL())
	assert.Equal(t, 90 * time.Second, g[p].Children()[0].(*graph.OriginNode).MinTTL())
}

func TestBuildGraphs_MedianAndOriginTTL(t *testing.T) {
	f, err1 := ParseJSON([]byte(`
		{
		  "pricemodels": {
			"A/B": {
			  "method": "median",
			  "ttl": 300,
			  "sources": [
				[{"origin": "ab", "pair": "A/B", "ttl": 120}]
			  ]
			}
		  }
		}
	`))
	assert.Nil(t, err1)

	p, _ := graph.NewPair("A/B")
	g, _ := f.BuildGraphs()

	// TTL assigned TTL should have higher priority:
	assert.Equal(t, 120 * time.Second, g[p].Children()[0].(*graph.OriginNode).MaxTTL())
	assert.Equal(t, 90 * time.Second, g[p].Children()[0].(*graph.OriginNode).MinTTL())
}
