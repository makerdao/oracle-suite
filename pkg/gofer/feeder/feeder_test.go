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

package feeder

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/makerdao/gofer/pkg/gofer/graph"
	"github.com/makerdao/gofer/pkg/gofer/origins"
	"github.com/makerdao/gofer/pkg/log/null"
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
	f := NewFeeder(originsSetMock(nil), null.New())

	// Feed method shouldn't panic
	warns := f.Feed([]graph.Node{})

	assert.Len(t, warns.List, 0)
}

func TestFeeder_Feed_NoFeedableNodes(t *testing.T) {
	f := NewFeeder(originsSetMock(nil), null.New())
	g := graph.NewMedianAggregatorNode(graph.Pair{Base: "A", Quote: "B"}, 1)

	// Feed method shouldn't panic
	warns := f.Feed([]graph.Node{g})

	assert.Len(t, warns.List, 0)
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

	f := NewFeeder(s, null.New())

	g := graph.NewMedianAggregatorNode(graph.Pair{Base: "A", Quote: "B"}, 1)
	o := graph.NewOriginNode(graph.OriginPair{
		Origin: "test",
		Pair:   graph.Pair{Base: "A", Quote: "B"},
	}, 0, 0)

	g.AddChild(o)
	warns := f.Feed([]graph.Node{g})

	assert.Len(t, warns.List, 0)
	assert.Equal(t, graph.Pair{Base: "A", Quote: "B"}, o.Tick().Pair)
	assert.Equal(t, 10.0, o.Tick().Price)
	assert.Equal(t, 9.0, o.Tick().Bid)
	assert.Equal(t, 11.0, o.Tick().Ask)
	assert.Equal(t, 10.0, o.Tick().Volume24h)
	assert.Equal(t, time.Unix(10000, 0), o.Tick().Timestamp)
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

	f := NewFeeder(s, null.New())

	g := graph.NewMedianAggregatorNode(graph.Pair{Base: "A", Quote: "B"}, 1)
	o1 := graph.NewOriginNode(graph.OriginPair{
		Origin: "test",
		Pair:   graph.Pair{Base: "A", Quote: "B"},
	}, 0, 0)
	o2 := graph.NewOriginNode(graph.OriginPair{
		Origin: "test",
		Pair:   graph.Pair{Base: "C", Quote: "D"},
	}, 0, 0)
	o3 := graph.NewOriginNode(graph.OriginPair{
		Origin: "test2",
		Pair:   graph.Pair{Base: "E", Quote: "F"},
	}, 0, 0)
	o4 := graph.NewOriginNode(graph.OriginPair{
		Origin: "test2",
		Pair:   graph.Pair{Base: "E", Quote: "F"},
	}, 0, 0)

	// The last o4 origin is intentionally same as an o3 origin. Also an o3
	// origin was added two times as a child for the g node. The feeder should
	// ask for E/F pair only once.

	g.AddChild(o1)
	g.AddChild(o2)
	g.AddChild(o3)
	g.AddChild(o3) // intentionally
	g.AddChild(o4)
	warns := f.Feed([]graph.Node{g})

	assert.Len(t, warns.List, 0)

	assert.Equal(t, graph.Pair{Base: "A", Quote: "B"}, o1.Tick().Pair)
	assert.Equal(t, 10.0, o1.Tick().Price)
	assert.Equal(t, 9.0, o1.Tick().Bid)
	assert.Equal(t, 11.0, o1.Tick().Ask)
	assert.Equal(t, 10.0, o1.Tick().Volume24h)
	assert.Equal(t, time.Unix(10000, 0), o1.Tick().Timestamp)

	assert.Equal(t, graph.Pair{Base: "C", Quote: "D"}, o2.Tick().Pair)
	assert.Equal(t, 20.0, o2.Tick().Price)
	assert.Equal(t, 19.0, o2.Tick().Bid)
	assert.Equal(t, 21.0, o2.Tick().Ask)
	assert.Equal(t, 20.0, o2.Tick().Volume24h)
	assert.Equal(t, time.Unix(20000, 0), o2.Tick().Timestamp)

	assert.Equal(t, graph.Pair{Base: "E", Quote: "F"}, o3.Tick().Pair)
	assert.Equal(t, 30.0, o3.Tick().Price)
	assert.Equal(t, 39.0, o3.Tick().Bid)
	assert.Equal(t, 31.0, o3.Tick().Ask)
	assert.Equal(t, 30.0, o3.Tick().Volume24h)
	assert.Equal(t, time.Unix(30000, 0), o3.Tick().Timestamp)

	assert.Equal(t, graph.Pair{Base: "E", Quote: "F"}, o4.Tick().Pair)
	assert.Equal(t, 30.0, o4.Tick().Price)
	assert.Equal(t, 39.0, o4.Tick().Bid)
	assert.Equal(t, 31.0, o4.Tick().Ask)
	assert.Equal(t, 30.0, o4.Tick().Volume24h)
	assert.Equal(t, time.Unix(30000, 0), o4.Tick().Timestamp)

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

	f := NewFeeder(s, null.New())

	g := graph.NewMedianAggregatorNode(graph.Pair{Base: "A", Quote: "B"}, 1)
	i := graph.NewIndirectAggregatorNode(graph.Pair{Base: "A", Quote: "B"})
	o := graph.NewOriginNode(graph.OriginPair{
		Origin: "test",
		Pair:   graph.Pair{Base: "A", Quote: "B"},
	}, 0, 0)

	g.AddChild(i)
	i.AddChild(o)
	warns := f.Feed([]graph.Node{g})

	assert.Len(t, warns.List, 0)
	assert.Equal(t, graph.Pair{Base: "A", Quote: "B"}, o.Tick().Pair)
	assert.Equal(t, 10.0, o.Tick().Price)
	assert.Equal(t, 9.0, o.Tick().Bid)
	assert.Equal(t, 11.0, o.Tick().Ask)
	assert.Equal(t, 10.0, o.Tick().Volume24h)
	assert.Equal(t, time.Unix(10000, 0), o.Tick().Timestamp)
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

	f := NewFeeder(s, null.New())

	g := graph.NewMedianAggregatorNode(graph.Pair{Base: "A", Quote: "B"}, 1)
	o := graph.NewOriginNode(graph.OriginPair{
		Origin: "test",
		Pair:   graph.Pair{Base: "A", Quote: "B"},
	}, 10*time.Second, 10*time.Second)

	_ = o.Ingest(graph.OriginTick{
		Tick: graph.Tick{
			Pair:      graph.Pair{Base: "A", Quote: "B"},
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
	warns := f.Feed([]graph.Node{g})

	// OriginNode shouldn't be updated because time diff is below MinTTL setting:
	assert.Len(t, warns.List, 0)
	assert.Equal(t, graph.Pair{Base: "A", Quote: "B"}, o.Tick().Pair)
	assert.Equal(t, 10.0, o.Tick().Price)
	assert.Equal(t, 9.0, o.Tick().Bid)
	assert.Equal(t, 11.0, o.Tick().Ask)
	assert.Equal(t, 10.0, o.Tick().Volume24h)
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

	f := NewFeeder(s, null.New())

	g := graph.NewMedianAggregatorNode(graph.Pair{Base: "A", Quote: "B"}, 1)
	o := graph.NewOriginNode(graph.OriginPair{
		Origin: "test",
		Pair:   graph.Pair{Base: "A", Quote: "B"},
	}, 10*time.Second, 60*time.Second)

	_ = o.Ingest(graph.OriginTick{
		Tick: graph.Tick{
			Pair:      graph.Pair{Base: "A", Quote: "B"},
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
	warns := f.Feed([]graph.Node{g})

	// OriginNode should be updated because time diff is above MinTTL setting:
	assert.Len(t, warns.List, 0)
	assert.Equal(t, graph.Pair{Base: "A", Quote: "B"}, o.Tick().Pair)
	assert.Equal(t, 11.0, o.Tick().Price)
	assert.Equal(t, 10.0, o.Tick().Bid)
	assert.Equal(t, 12.0, o.Tick().Ask)
	assert.Equal(t, 11.0, o.Tick().Volume24h)
}

func Test_getMinTTL(t *testing.T) {
	p := graph.Pair{Base: "A", Quote: "B"}
	root := graph.NewMedianAggregatorNode(p, 1)
	ttl := time.Second * time.Duration(time.Now().Unix()+10)
	on1 := graph.NewOriginNode(graph.OriginPair{Origin: "a", Pair: p}, 12*time.Second, ttl)
	on2 := graph.NewOriginNode(graph.OriginPair{Origin: "b", Pair: p}, 5*time.Second, ttl)
	on3 := graph.NewOriginNode(graph.OriginPair{Origin: "b", Pair: p}, 10*time.Second, ttl)

	root.AddChild(on1)
	root.AddChild(on2)
	root.AddChild(on3)

	assert.Equal(t, 5*time.Second, getMinTTL([]graph.Node{root}))
}

func Test_getMinTTL_SorterThanOneSecond(t *testing.T) {
	p := graph.Pair{Base: "A", Quote: "B"}
	root := graph.NewMedianAggregatorNode(p, 1)
	ttl := time.Second * time.Duration(time.Now().Unix()+10)
	on1 := graph.NewOriginNode(graph.OriginPair{Origin: "a", Pair: p}, 12*time.Second, ttl)
	on2 := graph.NewOriginNode(graph.OriginPair{Origin: "b", Pair: p}, -5*time.Second, ttl)
	on3 := graph.NewOriginNode(graph.OriginPair{Origin: "b", Pair: p}, 0*time.Second, ttl)

	root.AddChild(on1)
	root.AddChild(on2)
	root.AddChild(on3)

	assert.Equal(t, 1*time.Second, getMinTTL([]graph.Node{root}))
}
