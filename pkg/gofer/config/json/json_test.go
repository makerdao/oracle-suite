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

package json

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/makerdao/gofer/pkg/gofer/config"
	"github.com/makerdao/gofer/pkg/gofer/graph"
)

func TestParseJSONFile_ValidConfig(t *testing.T) {
	err := ParseJSONFile(&config.Config{}, "./testdata/config.valid.json")
	assert.NoError(t, err)
}

func TestParseJSONFile_MissingFile(t *testing.T) {
	err := ParseJSONFile(&config.Config{}, "missing")
	assert.Error(t, err)
}

func TestBuildGraphs_CyclicConfig(t *testing.T) {
	c := &config.Config{}
	err1 := ParseJSONFile(c, "./testdata/config.cyclic.json")
	assert.Nil(t, err1)

	_, err2 := c.BuildGraphs()
	assert.Error(t, err2)
}

func TestBuildGraphs_NoSources(t *testing.T) {
	// Price model without sources should be parsed without an error:
	c := &config.Config{}
	err1 := ParseJSON(c, []byte(`
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

	_, err2 := c.BuildGraphs()
	assert.Nil(t, err2)
}

func TestBuildGraphs_MissingParams(t *testing.T) {
	// The "params" may be skipped:
	c := &config.Config{}
	err1 := ParseJSON(c, []byte(`
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

	_, err2 := c.BuildGraphs()
	assert.Nil(t, err2)
}

func TestBuildGraphs_InvalidPairName(t *testing.T) {
	// The "A_B" is incorrect, "A/B" should be used instead:
	c := &config.Config{}
	err1 := ParseJSON(c, []byte(`
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

	_, err2 := c.BuildGraphs()
	assert.Error(t, err2)
}

func TestBuildGraphs_ReferenceToMissingPair(t *testing.T) {
	c := &config.Config{}
	err1 := ParseJSON(c, []byte(`
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

	_, err2 := c.BuildGraphs()
	assert.Error(t, err2)
}

func TestBuildGraphs_ReferenceToSelf(t *testing.T) {
	c := &config.Config{}
	err1 := ParseJSON(c, []byte(`
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

	_, err2 := c.BuildGraphs()
	assert.Error(t, err2)
}

func TestBuildGraphs_DefaultTTL(t *testing.T) {
	c := &config.Config{}
	err1 := ParseJSON(c, []byte(`
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
	g, _ := c.BuildGraphs()

	assert.Equal(t, 60*time.Second, g[p].Children()[0].(*graph.OriginNode).MaxTTL())
	assert.Equal(t, 30*time.Second, g[p].Children()[0].(*graph.OriginNode).MinTTL())
}

func TestBuildGraphs_OriginTTL(t *testing.T) {
	c := &config.Config{}
	err1 := ParseJSON(c, []byte(`
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
	g, _ := c.BuildGraphs()

	assert.Equal(t, 120*time.Second, g[p].Children()[0].(*graph.OriginNode).MaxTTL())
	assert.Equal(t, 90*time.Second, g[p].Children()[0].(*graph.OriginNode).MinTTL())
}

func TestBuildGraphs_MedianTTL(t *testing.T) {
	c := &config.Config{}
	err1 := ParseJSON(c, []byte(`
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
	g, _ := c.BuildGraphs()

	assert.Equal(t, 120*time.Second, g[p].Children()[0].(*graph.OriginNode).MaxTTL())
	assert.Equal(t, 90*time.Second, g[p].Children()[0].(*graph.OriginNode).MinTTL())
}

func TestBuildGraphs_MedianAndOriginTTL(t *testing.T) {
	c := &config.Config{}
	err1 := ParseJSON(c, []byte(`
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
	g, _ := c.BuildGraphs()

	// TTL assigned TTL should have higher priority:
	assert.Equal(t, 120*time.Second, g[p].Children()[0].(*graph.OriginNode).MaxTTL())
	assert.Equal(t, 90*time.Second, g[p].Children()[0].(*graph.OriginNode).MinTTL())
}
