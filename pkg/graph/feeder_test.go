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

package graph

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/makerdao/gofer/pkg/origins"
)

type mockHandler struct {
	// list of mocked Ticks to be returned by Fetch method
	mockedTicks map[origins.Pair]origins.Tick
	// list of pairs provided to Fetch method on last call
	fetchPairs []origins.Pair
}

func (m *mockHandler) Fetch(pairs []origins.Pair) []origins.FetchResult {
	m.fetchPairs = pairs

	var fr []origins.FetchResult
	for _, pair := range pairs {
		fr = append(fr, origins.FetchResult{
			Tick:  m.mockedTicks[pair],
			Error: nil,
		})
	}

	return fr
}

func originsSetMock(ticks map[string][]origins.Tick) *origins.Set {
	handlers := map[string]origins.Handler{}
	for origin, ticks := range ticks {
		ticksMap := map[origins.Pair]origins.Tick{}
		for _, tick := range ticks {
			ticksMap[tick.Pair] = tick
		}

		handlers[origin] = &mockHandler{mockedTicks: ticksMap}
	}

	return origins.NewSet(handlers)
}

func TestFeeder_Feed_EmptyGraph(t *testing.T) {
	f := NewFeeder(originsSetMock(nil))
	err := f.UpdateNodes([]Node{})

	assert.NoError(t, err)

	// Feed method shouldn't panic
}

func TestFeeder_Feed_NoFeedableNodes(t *testing.T) {
	f := NewFeeder(originsSetMock(nil))
	g := NewMedianAggregatorNode(Pair{Base: "A", Quote: "B"}, 1)
	err := f.UpdateNodes([]Node{g})

	assert.NoError(t, err)

	// Feed method shouldn't panic
}

func TestFeeder_Feed_OneOriginNode(t *testing.T) {
	s := originsSetMock(map[string][]origins.Tick{
		"test": {
			origins.Tick{
				Pair:      origins.Pair{Base: "A", Quote: "B"},
				Price:     10,
				Bid:       9,
				Ask:       11,
				Volume24h: 10,
				Timestamp: time.Unix(10000, 0),
			},
		},
	})

	f := NewFeeder(s)

	g := NewMedianAggregatorNode(Pair{Base: "A", Quote: "B"}, 1)
	o := NewOriginNode(OriginPair{
		Origin: "test",
		Pair:   Pair{Base: "A", Quote: "B"},
	}, 0, 0)

	g.AddChild(o)
	err := f.UpdateNodes([]Node{g})

	assert.NoError(t, err)
	assert.Equal(t, Pair{Base: "A", Quote: "B"}, o.tick.Pair)
	assert.Equal(t, 10.0, o.tick.Price)
	assert.Equal(t, 9.0, o.tick.Bid)
	assert.Equal(t, 11.0, o.tick.Ask)
	assert.Equal(t, 10.0, o.tick.Volume24h)
	assert.Equal(t, time.Unix(10000, 0), o.tick.Timestamp)
}

func TestFeeder_Feed_ManyOriginNodes(t *testing.T) {
	s := originsSetMock(map[string][]origins.Tick{
		"test": {
			origins.Tick{
				Pair:      origins.Pair{Base: "A", Quote: "B"},
				Price:     10,
				Bid:       9,
				Ask:       11,
				Volume24h: 10,
				Timestamp: time.Unix(10000, 0),
			},
			origins.Tick{
				Pair:      origins.Pair{Base: "C", Quote: "D"},
				Price:     20,
				Bid:       19,
				Ask:       21,
				Volume24h: 20,
				Timestamp: time.Unix(20000, 0),
			},
		},
		"test2": {
			origins.Tick{
				Pair:      origins.Pair{Base: "E", Quote: "F"},
				Price:     30,
				Bid:       39,
				Ask:       31,
				Volume24h: 30,
				Timestamp: time.Unix(30000, 0),
			},
		},
	})

	f := NewFeeder(s)

	g := NewMedianAggregatorNode(Pair{Base: "A", Quote: "B"}, 1)
	o1 := NewOriginNode(OriginPair{
		Origin: "test",
		Pair:   Pair{Base: "A", Quote: "B"},
	}, 0, 0)
	o2 := NewOriginNode(OriginPair{
		Origin: "test",
		Pair:   Pair{Base: "C", Quote: "D"},
	}, 0, 0)
	o3 := NewOriginNode(OriginPair{
		Origin: "test2",
		Pair:   Pair{Base: "E", Quote: "F"},
	}, 0, 0)
	o4 := NewOriginNode(OriginPair{
		Origin: "test2",
		Pair:   Pair{Base: "E", Quote: "F"},
	}, 0, 0)

	// The last o4 origin is intentionally same as an o3 origin. Also an o3
	// origin was added two times as a child for the g node. The feeder should
	// ask for E/F pair only once.

	g.AddChild(o1)
	g.AddChild(o2)
	g.AddChild(o3)
	g.AddChild(o3) // intentionally
	g.AddChild(o4)
	err := f.UpdateNodes([]Node{g})

	assert.NoError(t, err)

	assert.Equal(t, Pair{Base: "A", Quote: "B"}, o1.tick.Pair)
	assert.Equal(t, 10.0, o1.tick.Price)
	assert.Equal(t, 9.0, o1.tick.Bid)
	assert.Equal(t, 11.0, o1.tick.Ask)
	assert.Equal(t, 10.0, o1.tick.Volume24h)
	assert.Equal(t, time.Unix(10000, 0), o1.tick.Timestamp)

	assert.Equal(t, Pair{Base: "C", Quote: "D"}, o2.tick.Pair)
	assert.Equal(t, 20.0, o2.tick.Price)
	assert.Equal(t, 19.0, o2.tick.Bid)
	assert.Equal(t, 21.0, o2.tick.Ask)
	assert.Equal(t, 20.0, o2.tick.Volume24h)
	assert.Equal(t, time.Unix(20000, 0), o2.tick.Timestamp)

	assert.Equal(t, Pair{Base: "E", Quote: "F"}, o3.tick.Pair)
	assert.Equal(t, 30.0, o3.tick.Price)
	assert.Equal(t, 39.0, o3.tick.Bid)
	assert.Equal(t, 31.0, o3.tick.Ask)
	assert.Equal(t, 30.0, o3.tick.Volume24h)
	assert.Equal(t, time.Unix(30000, 0), o3.tick.Timestamp)

	assert.Equal(t, Pair{Base: "E", Quote: "F"}, o4.tick.Pair)
	assert.Equal(t, 30.0, o4.tick.Price)
	assert.Equal(t, 39.0, o4.tick.Bid)
	assert.Equal(t, 31.0, o4.tick.Ask)
	assert.Equal(t, 30.0, o4.tick.Volume24h)
	assert.Equal(t, time.Unix(30000, 0), o4.tick.Timestamp)

	// Check if pairs was properly grouped per origins and check if the E/F pair
	// appeared only once:
	testPairs := s.Handlers()["test"].(*mockHandler).fetchPairs
	test2Pairs := s.Handlers()["test2"].(*mockHandler).fetchPairs
	assert.ElementsMatch(t, []origins.Pair{{Base: "A", Quote: "B"}, {Base: "C", Quote: "D"}}, testPairs)
	assert.ElementsMatch(t, []origins.Pair{{Base: "E", Quote: "F"}}, test2Pairs)
}

func TestFeeder_Feed_NestedOriginNode(t *testing.T) {
	s := originsSetMock(map[string][]origins.Tick{
		"test": {
			origins.Tick{
				Pair:      origins.Pair{Base: "A", Quote: "B"},
				Price:     10,
				Bid:       9,
				Ask:       11,
				Volume24h: 10,
				Timestamp: time.Unix(10000, 0),
			},
		},
	})

	f := NewFeeder(s)

	g := NewMedianAggregatorNode(Pair{Base: "A", Quote: "B"}, 1)
	i := NewIndirectAggregatorNode(Pair{Base: "A", Quote: "B"})
	o := NewOriginNode(OriginPair{
		Origin: "test",
		Pair:   Pair{Base: "A", Quote: "B"},
	}, 0, 0)

	g.AddChild(i)
	i.AddChild(o)
	err := f.UpdateNodes([]Node{g})

	assert.NoError(t, err)
	assert.Equal(t, Pair{Base: "A", Quote: "B"}, o.tick.Pair)
	assert.Equal(t, 10.0, o.tick.Price)
	assert.Equal(t, 9.0, o.tick.Bid)
	assert.Equal(t, 11.0, o.tick.Ask)
	assert.Equal(t, 10.0, o.tick.Volume24h)
	assert.Equal(t, time.Unix(10000, 0), o.tick.Timestamp)
}

func TestFeeder_Feed_BelowMinTTL(t *testing.T) {
	s := originsSetMock(map[string][]origins.Tick{
		"test": {
			origins.Tick{
				Pair:      origins.Pair{Base: "A", Quote: "B"},
				Price:     11,
				Bid:       10,
				Ask:       12,
				Volume24h: 11,
				Timestamp: time.Unix(10000, 0),
			},
		},
	})

	f := NewFeeder(s)

	g := NewMedianAggregatorNode(Pair{Base: "A", Quote: "B"}, 1)
	o := NewOriginNode(OriginPair{
		Origin: "test",
		Pair:   Pair{Base: "A", Quote: "B"},
	}, 10*time.Second, 10*time.Second)

	_ = o.Ingest(OriginTick{
		Tick: Tick{
			Pair:      Pair{Base: "A", Quote: "B"},
			Price:     10,
			Bid:       9,
			Ask:       11,
			Volume24h: 10,
			Timestamp: time.Now().Add(-5 * time.Second),
		},
		Origin: "test",
		Error:  nil,
	})

	g.AddChild(o)
	err := f.UpdateNodes([]Node{g})

	// OriginNode shouldn't be updated because time diff is below MinTTL setting:
	assert.NoError(t, err)
	assert.Equal(t, Pair{Base: "A", Quote: "B"}, o.tick.Pair)
	assert.Equal(t, 10.0, o.tick.Price)
	assert.Equal(t, 9.0, o.tick.Bid)
	assert.Equal(t, 11.0, o.tick.Ask)
	assert.Equal(t, 10.0, o.tick.Volume24h)
}

func TestFeeder_Feed_BetweenTTLs(t *testing.T) {
	s := originsSetMock(map[string][]origins.Tick{
		"test": {
			origins.Tick{
				Pair:      origins.Pair{Base: "A", Quote: "B"},
				Price:     11,
				Bid:       10,
				Ask:       12,
				Volume24h: 11,
				Timestamp: time.Unix(10000, 0),
			},
		},
	})

	f := NewFeeder(s)

	g := NewMedianAggregatorNode(Pair{Base: "A", Quote: "B"}, 1)
	o := NewOriginNode(OriginPair{
		Origin: "test",
		Pair:   Pair{Base: "A", Quote: "B"},
	}, 10*time.Second, 60*time.Second)

	_ = o.Ingest(OriginTick{
		Tick: Tick{
			Pair:      Pair{Base: "A", Quote: "B"},
			Price:     10,
			Bid:       9,
			Ask:       11,
			Volume24h: 10,
			Timestamp: time.Now().Add(-30 * time.Second),
		},
		Origin: "test",
		Error:  nil,
	})

	g.AddChild(o)
	err := f.UpdateNodes([]Node{g})

	// OriginNode should be updated because time diff is above MinTTL setting:
	assert.NoError(t, err)
	assert.Equal(t, Pair{Base: "A", Quote: "B"}, o.tick.Pair)
	assert.Equal(t, 11.0, o.tick.Price)
	assert.Equal(t, 10.0, o.tick.Bid)
	assert.Equal(t, 12.0, o.tick.Ask)
	assert.Equal(t, 11.0, o.tick.Volume24h)
}
