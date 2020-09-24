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
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMedianAggregatorNode_Children(t *testing.T) {
	m := NewMedianAggregatorNode(Pair{Base: "A", Quote: "B"}, 3)

	c1 := NewOriginNode(OriginPair{Pair: Pair{Base: "A", Quote: "B"}, Origin: "a"})
	c2 := NewOriginNode(OriginPair{Pair: Pair{Base: "A", Quote: "B"}, Origin: "b"})
	c3 := NewOriginNode(OriginPair{Pair: Pair{Base: "A", Quote: "B"}, Origin: "c"})

	m.AddChild(c1)
	m.AddChild(c2)
	m.AddChild(c3)

	assert.Len(t, m.Children(), 3)
	assert.Same(t, c1, m.Children()[0])
	assert.Same(t, c2, m.Children()[1])
	assert.Same(t, c3, m.Children()[2])
}

func TestMedianAggregatorNode_Pair(t *testing.T) {
	p := Pair{Base: "A", Quote: "B"}
	m := NewMedianAggregatorNode(p, 3)

	assert.Equal(t, m.Pair(), p)
}

func TestMedianAggregatorNode_Tick_ThreeOriginTicks(t *testing.T) {
	p := Pair{Base: "A", Quote: "B"}
	n := time.Now()
	m := NewMedianAggregatorNode(p, 3)

	c1 := NewOriginNode(OriginPair{Pair: p, Origin: "a"})
	c2 := NewOriginNode(OriginPair{Pair: p, Origin: "b"})
	c3 := NewOriginNode(OriginPair{Pair: p, Origin: "c"})

	c1.Ingest(OriginTick{
		Tick: Tick{
			Pair:      p,
			Price:     10,
			Bid:       10,
			Ask:       10,
			Volume24h: 10,
			Timestamp: n,
		},
		Origin: "a",
		Error:  nil,
	})

	c2.Ingest(OriginTick{
		Tick: Tick{
			Pair:      p,
			Price:     20,
			Bid:       20,
			Ask:       20,
			Volume24h: 20,
			Timestamp: n,
		},
		Origin: "b",
		Error:  nil,
	})

	c3.Ingest(OriginTick{
		Tick: Tick{
			Pair:      p,
			Price:     30,
			Bid:       30,
			Ask:       30,
			Volume24h: 30,
			Timestamp: n,
		},
		Origin: "c",
		Error:  nil,
	})

	m.AddChild(c1)
	m.AddChild(c2)
	m.AddChild(c3)

	expected := AggregatorTick{
		Tick: Tick{
			Pair:      p,
			Price:     20,
			Bid:       20,
			Ask:       20,
			Volume24h: 0,
			Timestamp: n,
		},
		OriginTicks:     []OriginTick{c1.Tick(), c2.Tick(), c3.Tick()},
		AggregatorTicks: nil,
		Parameters:      map[string]string{"method": "median", "min": "3"},
		Error:           nil,
	}

	assert.Equal(t, expected, m.Tick())
}

func TestMedianAggregatorNode_Tick_ThreeAggregatorTicks(t *testing.T) {
	p := Pair{Base: "A", Quote: "B"}
	n := time.Now()
	m := NewMedianAggregatorNode(p, 3)

	c1 := NewOriginNode(OriginPair{Pair: p, Origin: "a"})
	c2 := NewOriginNode(OriginPair{Pair: p, Origin: "b"})
	c3 := NewOriginNode(OriginPair{Pair: p, Origin: "c"})

	i1 := NewMedianAggregatorNode(p, 1)
	i1.AddChild(c1)

	i2 := NewMedianAggregatorNode(p, 1)
	i2.AddChild(c2)

	i3 := NewMedianAggregatorNode(p, 1)
	i3.AddChild(c3)

	c1.Ingest(OriginTick{
		Tick: Tick{
			Pair:      p,
			Price:     10,
			Bid:       10,
			Ask:       10,
			Volume24h: 10,
			Timestamp: n,
		},
		Origin: "a",
		Error:  nil,
	})

	c2.Ingest(OriginTick{
		Tick: Tick{
			Pair:      p,
			Price:     20,
			Bid:       20,
			Ask:       20,
			Volume24h: 20,
			Timestamp: n,
		},
		Origin: "b",
		Error:  nil,
	})

	c3.Ingest(OriginTick{
		Tick: Tick{
			Pair:      p,
			Price:     30,
			Bid:       30,
			Ask:       30,
			Volume24h: 30,
			Timestamp: n,
		},
		Origin: "c",
		Error:  nil,
	})

	m.AddChild(i1)
	m.AddChild(i2)
	m.AddChild(i3)

	expected := AggregatorTick{
		Tick: Tick{
			Pair:      p,
			Price:     20,
			Bid:       20,
			Ask:       20,
			Volume24h: 0,
			Timestamp: n,
		},
		OriginTicks: nil,
		AggregatorTicks: []AggregatorTick{
			{
				Tick: Tick{
					Pair:      p,
					Price:     10,
					Bid:       10,
					Ask:       10,
					Volume24h: 0,
					Timestamp: n,
				},
				OriginTicks:     []OriginTick{c1.Tick()},
				AggregatorTicks: nil,
				Parameters:      map[string]string{"method": "median", "min": "1"},
				Error:           nil,
			},
			{
				Tick: Tick{
					Pair:      p,
					Price:     20,
					Bid:       20,
					Ask:       20,
					Volume24h: 0,
					Timestamp: n,
				},
				OriginTicks:     []OriginTick{c2.Tick()},
				AggregatorTicks: nil,
				Parameters:      map[string]string{"method": "median", "min": "1"},
				Error:           nil,
			},
			{
				Tick: Tick{
					Pair:      p,
					Price:     30,
					Bid:       30,
					Ask:       30,
					Volume24h: 0,
					Timestamp: n,
				},
				OriginTicks:     []OriginTick{c3.Tick()},
				AggregatorTicks: nil,
				Parameters:      map[string]string{"method": "median", "min": "1"},
				Error:           nil,
			},
		},
		Parameters: map[string]string{"method": "median", "min": "3"},
		Error:      nil,
	}

	assert.Equal(t, expected, m.Tick())
}

func TestMedianAggregatorNode_Tick_NotEnoughSources(t *testing.T) {
	p := Pair{Base: "A", Quote: "B"}
	n := time.Now()
	m := NewMedianAggregatorNode(p, 3)

	c1 := NewOriginNode(OriginPair{Pair: p, Origin: "a"})

	c1.Ingest(OriginTick{
		Tick: Tick{
			Pair:      p,
			Price:     10,
			Bid:       10,
			Ask:       10,
			Volume24h: 10,
			Timestamp: n,
		},
		Origin: "a",
		Error:  nil,
	})

	m.AddChild(c1)

	tick := m.Tick()

	assert.True(t, errors.As(tick.Error, &NotEnoughSources{}))

	// If possible, the median should be calculated for the rest of the prices:
	assert.Equal(t, float64(10), tick.Price)
	assert.Equal(t, float64(10), tick.Bid)
	assert.Equal(t, float64(10), tick.Ask)
}

func TestMedianAggregatorNode_Tick_ChildTickWithError(t *testing.T) {
	p := Pair{Base: "A", Quote: "B"}
	n := time.Now()
	m := NewMedianAggregatorNode(p, 2)

	c1 := NewOriginNode(OriginPair{Pair: p, Origin: "a"})
	c2 := NewOriginNode(OriginPair{Pair: p, Origin: "b"})

	c1.Ingest(OriginTick{
		Tick: Tick{
			Pair:      p,
			Price:     10,
			Bid:       10,
			Ask:       10,
			Volume24h: 10,
			Timestamp: n,
		},
		Origin: "a",
		Error:  nil,
	})

	c2.Ingest(OriginTick{
		Tick: Tick{
			Pair:      p,
			Price:     20,
			Bid:       20,
			Ask:       20,
			Volume24h: 20,
			Timestamp: n,
		},
		Origin: "b",
		Error:  errors.New("something"),
	})

	m.AddChild(c1)
	m.AddChild(c2)

	tick := m.Tick()

	assert.True(t, errors.As(tick.Error, &NotEnoughSources{}))

	// If possible, the median should be calculated for the rest of the prices:
	assert.Equal(t, float64(10), tick.Price)
	assert.Equal(t, float64(10), tick.Bid)
	assert.Equal(t, float64(10), tick.Ask)
}

func TestMedianAggregatorNode_Tick_IncompatiblePairs(t *testing.T) {
	p1 := Pair{Base: "A", Quote: "B"}
	p2 := Pair{Base: "C", Quote: "D"}
	n := time.Now()
	m := NewMedianAggregatorNode(p1, 2)

	c1 := NewOriginNode(OriginPair{Pair: p1, Origin: "a"})
	c2 := NewOriginNode(OriginPair{Pair: p2, Origin: "b"})

	c1.Ingest(OriginTick{
		Tick: Tick{
			Pair:      p1,
			Price:     10,
			Bid:       10,
			Ask:       10,
			Volume24h: 10,
			Timestamp: n,
		},
		Origin: "a",
		Error:  nil,
	})

	c2.Ingest(OriginTick{
		Tick: Tick{
			Pair:      p2,
			Price:     20,
			Bid:       20,
			Ask:       20,
			Volume24h: 20,
			Timestamp: n,
		},
		Origin: "b",
		Error:  nil,
	})

	m.AddChild(c1)
	m.AddChild(c2)

	tick := m.Tick()

	assert.True(t, errors.As(tick.Error, &IncompatiblePairs{}))

	// If possible, the median should be calculated for the rest of the prices:
	assert.Equal(t, float64(10), tick.Price)
	assert.Equal(t, float64(10), tick.Bid)
	assert.Equal(t, float64(10), tick.Ask)
}

func TestMedianAggregatorNode_Tick_NoChildrenNodes(t *testing.T) {
	p := Pair{Base: "A", Quote: "B"}
	m := NewMedianAggregatorNode(p, 2)

	tick := m.Tick()

	assert.True(t, errors.As(tick.Error, &NotEnoughSources{}))
}

func TestMedianAggregatorNode_Tick_FilterOutPricesLteZero(t *testing.T) {
	p := Pair{Base: "A", Quote: "B"}
	n := time.Now()
	m := NewMedianAggregatorNode(p, 1)

	c1 := NewOriginNode(OriginPair{Pair: p, Origin: "a"})
	c2 := NewOriginNode(OriginPair{Pair: p, Origin: "b"})
	c3 := NewOriginNode(OriginPair{Pair: p, Origin: "c"})

	c1.Ingest(OriginTick{
		Tick: Tick{
			Pair:      p,
			Price:     10,
			Bid:       0,
			Ask:       0,
			Volume24h: 0,
			Timestamp: n,
		},
		Origin: "a",
		Error:  nil,
	})

	c2.Ingest(OriginTick{
		Tick: Tick{
			Pair:      p,
			Price:     0,
			Bid:       10,
			Ask:       0,
			Volume24h: 0,
			Timestamp: n,
		},
		Origin: "b",
		Error:  nil,
	})

	c3.Ingest(OriginTick{
		Tick: Tick{
			Pair:      p,
			Price:     0,
			Bid:       0,
			Ask:       10,
			Volume24h: 0,
			Timestamp: n,
		},
		Origin: "b",
		Error:  nil,
	})

	m.AddChild(c1)
	m.AddChild(c2)
	m.AddChild(c3)

	tick := m.Tick()

	assert.Equal(t, float64(10), tick.Price)
	assert.Equal(t, float64(10), tick.Bid)
	assert.Equal(t, float64(10), tick.Ask)
}

func Test_median(t *testing.T) {
	tests := []struct {
		name   string
		prices []float64
		want   float64
	}{
		{
			name:   "no-prices",
			prices: []float64{},
			want:   float64(0),
		},
		{
			name:   "one-price",
			prices: []float64{10},
			want:   float64(10),
		},
		{
			name:   "three-prices",
			prices: []float64{-20, 10, 20},
			want:   float64(10),
		},
		{
			name:   "four-prices",
			prices: []float64{10, 20, 30, 40},
			want:   float64(25),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := median(tt.prices)
			assert.Equal(t, tt.want, got)
		})
	}
}
