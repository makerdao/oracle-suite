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

package gofer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/makerdao/gofer/pkg/gofer/feeder"
	"github.com/makerdao/gofer/pkg/gofer/graph"
	"github.com/makerdao/gofer/pkg/gofer/origins"
	"github.com/makerdao/gofer/pkg/log/null"
)

var testGraph map[graph.Pair]graph.Aggregator
var testFeeder *feeder.Feeder
var testPairs = map[string]graph.Pair{
	"A/B": {Base: "A", Quote: "B"},
	"X/Y": {Base: "X", Quote: "Y"},
}

type testExchange struct{}

func (f *testExchange) Fetch(pairs []origins.Pair) []origins.FetchResult {
	var r []origins.FetchResult
	for _, p := range pairs {
		r = append(r, origins.FetchResult{
			Tick: origins.Tick{
				Pair:      p,
				Price:     10,
				Bid:       10,
				Ask:       10,
				Volume24h: 10,
				Timestamp: time.Unix(0, 0),
			},
			Error: nil,
		})
	}
	return r
}

func init() {
	ab := testPairs["A/B"]
	xy := testPairs["X/Y"]

	abGraph := graph.NewMedianAggregatorNode(ab, 0)
	abc1 := graph.NewOriginNode(graph.OriginPair{Origin: "a", Pair: ab}, 0, 0)
	abc2 := graph.NewOriginNode(graph.OriginPair{Origin: "b", Pair: ab}, 0, 0)
	abc3 := graph.NewMedianAggregatorNode(ab, 0)
	abGraph.AddChild(abc1)
	abGraph.AddChild(abc3)
	abc3.AddChild(abc1)
	abc3.AddChild(abc2)

	xyGraph := graph.NewMedianAggregatorNode(xy, 0)
	xyc1 := graph.NewOriginNode(graph.OriginPair{Origin: "x", Pair: xy}, 0, 0)
	xyc2 := graph.NewOriginNode(graph.OriginPair{Origin: "y", Pair: xy}, 0, 0)
	xyGraph.AddChild(xyc1)
	xyGraph.AddChild(xyc2)

	testGraph = map[graph.Pair]graph.Aggregator{
		ab: abGraph,
		xy: xyGraph,
	}

	testFeeder = feeder.NewFeeder(origins.NewSet(map[string]origins.Handler{
		"a": &testExchange{},
		"b": &testExchange{},
		"x": &testExchange{},
		"y": &testExchange{},
	}), null.New())
}

func TestGofer_Graphs(t *testing.T) {
	g := NewGofer(testGraph, testFeeder)
	assert.Equal(t, testGraph, g.Graphs())
}

func TestGofer_Feeder(t *testing.T) {
	g := NewGofer(testGraph, testFeeder)
	assert.Equal(t, testFeeder, g.Feeder())
}

func TestGofer_Pairs(t *testing.T) {
	g := NewGofer(testGraph, testFeeder)
	assert.ElementsMatch(t, []graph.Pair{
		testPairs["A/B"],
		testPairs["X/Y"],
	}, g.Pairs())
}

func TestGofer_Feed(t *testing.T) {
	g := NewGofer(testGraph, testFeeder)

	_, err := g.Feed(testPairs["A/B"])
	assert.NoError(t, err)
}

func TestGofer_Feed_MissingPair(t *testing.T) {
	g := NewGofer(testGraph, testFeeder)

	_, err := g.Feed(graph.Pair{Base: "C", Quote: "D"})
	assert.Error(t, err)
}

func TestGofer_Tick(t *testing.T) {
	g := NewGofer(testGraph, testFeeder)
	ab := testPairs["A/B"]

	tick, err := g.Tick(ab)
	assert.NoError(t, err)
	assert.Equal(t, ab, tick.Pair)
}

func TestGofer_Ticks(t *testing.T) {
	g := NewGofer(testGraph, testFeeder)
	ab := testPairs["A/B"]
	xy := testPairs["X/Y"]

	_, _ = g.Feed(ab, xy)
	ticks, err := g.Ticks(ab, xy)
	assert.NoError(t, err)
	assert.Equal(t, ab, ticks[0].Pair)
	assert.Equal(t, xy, ticks[1].Pair)
}

func TestGofer_Ticks_MissingPair(t *testing.T) {
	g := NewGofer(testGraph, testFeeder)

	_, err := g.Ticks(graph.Pair{Base: "C", Quote: "D"})
	assert.Error(t, err)
}

func TestGofer_Origins(t *testing.T) {
	g := NewGofer(testGraph, testFeeder)
	ab := testPairs["A/B"]
	xy := testPairs["X/Y"]

	list, err := g.Origins(ab, xy)
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"a", "b"}, list[ab])
	assert.ElementsMatch(t, []string{"x", "y"}, list[xy])
}

func TestGofer_Origins_MissingPair(t *testing.T) {
	g := NewGofer(testGraph, testFeeder)

	_, err := g.Origins(graph.Pair{Base: "C", Quote: "D"})
	assert.Error(t, err)
}
