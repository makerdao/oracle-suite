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

const testTTL = 10 * time.Second

func TestIndirectAggregatorNode_Children(t *testing.T) {
	m := NewIndirectAggregatorNode(Pair{Base: "A", Quote: "B"})

	c1 := NewOriginNode(OriginPair{Pair: Pair{Base: "A", Quote: "B"}, Origin: "a"}, testTTL, testTTL)
	c2 := NewOriginNode(OriginPair{Pair: Pair{Base: "A", Quote: "B"}, Origin: "b"}, testTTL, testTTL)
	c3 := NewOriginNode(OriginPair{Pair: Pair{Base: "A", Quote: "B"}, Origin: "c"}, testTTL, testTTL)

	m.AddChild(c1)
	m.AddChild(c2)
	m.AddChild(c3)

	assert.Len(t, m.Children(), 3)
	assert.Same(t, c1, m.Children()[0])
	assert.Same(t, c2, m.Children()[1])
	assert.Same(t, c3, m.Children()[2])
}

func TestIndirectAggregatorNode_Pair(t *testing.T) {
	p := Pair{Base: "A", Quote: "B"}
	m := NewIndirectAggregatorNode(p)

	assert.Equal(t, m.Pair(), p)
}

func TestIndirectAggregatorNode_Tick_ThreeOriginTicks(t *testing.T) {
	p1 := Pair{Base: "A", Quote: "B"}
	p2 := Pair{Base: "B", Quote: "C"}
	p3 := Pair{Base: "C", Quote: "D"}
	pf := Pair{Base: "A", Quote: "D"}

	n := time.Now()
	m := NewIndirectAggregatorNode(pf)

	c1 := NewOriginNode(OriginPair{Pair: p1, Origin: "a"}, testTTL, testTTL)
	c2 := NewOriginNode(OriginPair{Pair: p2, Origin: "b"}, testTTL, testTTL)
	c3 := NewOriginNode(OriginPair{Pair: p3, Origin: "c"}, testTTL, testTTL)

	_ = c1.Ingest(OriginTick{
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

	_ = c2.Ingest(OriginTick{
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

	_ = c3.Ingest(OriginTick{
		Tick: Tick{
			Pair:      p3,
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
			Pair:      pf,
			Price:     6000,
			Bid:       6000,
			Ask:       6000,
			Volume24h: 0,
			Timestamp: n,
		},
		OriginTicks:     []OriginTick{c1.Tick(), c2.Tick(), c3.Tick()},
		AggregatorTicks: nil,
		Parameters:      map[string]string{"method": "indirect"},
		Error:           nil,
	}

	assert.Equal(t, expected, m.Tick())
}

func TestIndirectAggregatorNode_Tick_ThreeAggregatorTicks(t *testing.T) {
	p1 := Pair{Base: "A", Quote: "B"}
	p2 := Pair{Base: "B", Quote: "C"}
	p3 := Pair{Base: "C", Quote: "D"}
	pf := Pair{Base: "A", Quote: "D"}

	n := time.Now()
	m := NewIndirectAggregatorNode(pf)

	c1 := NewOriginNode(OriginPair{Pair: p1, Origin: "a"}, testTTL, testTTL)
	c2 := NewOriginNode(OriginPair{Pair: p2, Origin: "b"}, testTTL, testTTL)
	c3 := NewOriginNode(OriginPair{Pair: p3, Origin: "c"}, testTTL, testTTL)

	i1 := NewIndirectAggregatorNode(p1)
	i1.AddChild(c1)

	i2 := NewIndirectAggregatorNode(p2)
	i2.AddChild(c2)

	i3 := NewIndirectAggregatorNode(p3)
	i3.AddChild(c3)

	_ = c1.Ingest(OriginTick{
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

	_ = c2.Ingest(OriginTick{
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

	_ = c3.Ingest(OriginTick{
		Tick: Tick{
			Pair:      p3,
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
			Pair:      pf,
			Price:     6000,
			Bid:       6000,
			Ask:       6000,
			Volume24h: 0,
			Timestamp: n,
		},
		OriginTicks: nil,
		AggregatorTicks: []AggregatorTick{
			{
				Tick: Tick{
					Pair:      p1,
					Price:     10,
					Bid:       10,
					Ask:       10,
					Volume24h: 10,
					Timestamp: n,
				},
				OriginTicks:     []OriginTick{c1.Tick()},
				AggregatorTicks: nil,
				Parameters:      map[string]string{"method": "indirect"},
				Error:           nil,
			},
			{
				Tick: Tick{
					Pair:      p2,
					Price:     20,
					Bid:       20,
					Ask:       20,
					Volume24h: 20,
					Timestamp: n,
				},
				OriginTicks:     []OriginTick{c2.Tick()},
				AggregatorTicks: nil,
				Parameters:      map[string]string{"method": "indirect"},
				Error:           nil,
			},
			{
				Tick: Tick{
					Pair:      p3,
					Price:     30,
					Bid:       30,
					Ask:       30,
					Volume24h: 30,
					Timestamp: n,
				},
				OriginTicks:     []OriginTick{c3.Tick()},
				AggregatorTicks: nil,
				Parameters:      map[string]string{"method": "indirect"},
				Error:           nil,
			},
		},
		Parameters: map[string]string{"method": "indirect"},
		Error:      nil,
	}

	assert.Equal(t, expected, m.Tick())
}

func TestIndirectAggregatorNode_Tick_ChildTickWithError(t *testing.T) {
	p1 := Pair{Base: "A", Quote: "B"}
	p2 := Pair{Base: "B", Quote: "C"}
	pf := Pair{Base: "A", Quote: "C"}

	n := time.Now()
	m := NewIndirectAggregatorNode(pf)

	c1 := NewOriginNode(OriginPair{Pair: p1, Origin: "a"}, testTTL, testTTL)
	c2 := NewOriginNode(OriginPair{Pair: p2, Origin: "b"}, testTTL, testTTL)

	_ = c1.Ingest(OriginTick{
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

	_ = c2.Ingest(OriginTick{
		Tick: Tick{
			Pair:      p2,
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

	assert.True(t, errors.As(m.Tick().Error, &ErrTick{}))
}

func TestIndirectAggregatorNode_Tick_ResolveToWrongPair(t *testing.T) {
	// Below pairs will be resolved resolve to the A/D but the A/C is expected:
	p1 := Pair{Base: "A", Quote: "B"}
	p2 := Pair{Base: "B", Quote: "D"}
	pf := Pair{Base: "A", Quote: "C"}

	n := time.Now()
	m := NewIndirectAggregatorNode(pf)

	c1 := NewOriginNode(OriginPair{Pair: p1, Origin: "a"}, testTTL, testTTL)
	c2 := NewOriginNode(OriginPair{Pair: p2, Origin: "b"}, testTTL, testTTL)

	_ = c1.Ingest(OriginTick{
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

	_ = c2.Ingest(OriginTick{
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

	assert.True(t, errors.As(m.Tick().Error, &ErrResolve{}))
}

func TestIndirectAggregatorNode_Tick_UnableToResolve(t *testing.T) {
	// It's impossible to resolve below pairs, because the A/B and C/D have no common part:
	p1 := Pair{Base: "A", Quote: "B"}
	p2 := Pair{Base: "C", Quote: "D"}
	pf := Pair{Base: "A", Quote: "C"}

	n := time.Now()
	m := NewIndirectAggregatorNode(pf)

	c1 := NewOriginNode(OriginPair{Pair: p1, Origin: "a"}, testTTL, testTTL)
	c2 := NewOriginNode(OriginPair{Pair: p2, Origin: "b"}, testTTL, testTTL)

	_ = c1.Ingest(OriginTick{
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

	_ = c2.Ingest(OriginTick{
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

	// Resolved to the A/D pair but the A/C pair was expected:
	assert.True(t, errors.As(m.Tick().Error, &ErrNoCommonPart{}))
}

func TestIndirectAggregatorNode_Tick_DivByZero(t *testing.T) {
	p1 := Pair{Base: "A", Quote: "B"}
	p2 := Pair{Base: "C", Quote: "B"}
	pf := Pair{Base: "A", Quote: "C"}

	n := time.Now()
	m := NewIndirectAggregatorNode(pf)

	c1 := NewOriginNode(OriginPair{Pair: p1, Origin: "a"}, testTTL, testTTL)
	c2 := NewOriginNode(OriginPair{Pair: p2, Origin: "b"}, testTTL, testTTL)

	_ = c1.Ingest(OriginTick{
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

	_ = c2.Ingest(OriginTick{
		Tick: Tick{
			Pair:      p2,
			Price:     0,
			Bid:       0,
			Ask:       0,
			Volume24h: 0,
			Timestamp: n,
		},
		Origin: "b",
		Error:  nil,
	})

	m.AddChild(c1)
	m.AddChild(c2)

	// Conversion between the A/B and the C/B requires division. The C/B pair
	// reports 0 as its price so the ErrDivByZero error should be returned:
	assert.True(t, errors.As(m.Tick().Error, &ErrDivByZero{}))
}

func Test_crossRate(t *testing.T) {
	tests := []struct {
		name    string
		ticks   []Tick
		want    Tick
		wantErr bool
	}{
		{
			name:    "no-pair",
			ticks:   []Tick{},
			want:    Tick{},
			wantErr: false,
		},
		{
			name: "one-pair",
			ticks: []Tick{
				{
					Pair:      Pair{Base: "A", Quote: "B"},
					Price:     10,
					Bid:       5,
					Ask:       15,
					Volume24h: 10,
					Timestamp: time.Unix(10, 0),
				},
			},
			want: Tick{
				Pair:      Pair{Base: "A", Quote: "B"},
				Price:     10,
				Bid:       5,
				Ask:       15,
				Volume24h: 10,
				Timestamp: time.Unix(10, 0),
			},
			wantErr: false,
		},
		{
			name: "invalid-conversion",
			ticks: []Tick{
				{Pair: Pair{Base: "A", Quote: "B"}},
				{Pair: Pair{Base: "X", Quote: "Y"}},
			},
			want:    Tick{},
			wantErr: true,
		},
		{
			name: "ac/bc",
			ticks: []Tick{
				{
					Pair:      Pair{Base: "A", Quote: "C"},
					Price:     10,
					Bid:       5,
					Ask:       15,
					Volume24h: 10,
					Timestamp: time.Unix(10, 0),
				},
				{
					Pair:      Pair{Base: "B", Quote: "C"},
					Price:     20,
					Bid:       10,
					Ask:       30,
					Volume24h: 10,
					Timestamp: time.Unix(10, 0),
				},
			},
			want: Tick{
				Pair:      Pair{Base: "A", Quote: "C"},
				Price:     float64(10) / 20,
				Bid:       float64(5) / 10,
				Ask:       float64(15) / 30,
				Volume24h: 10,
				Timestamp: time.Unix(10, 0),
			},
			wantErr: false,
		},
		{
			name: "ac/bc-divByZero",
			ticks: []Tick{
				{
					Pair:      Pair{Base: "A", Quote: "C"},
					Price:     10,
					Bid:       5,
					Ask:       15,
					Volume24h: 10,
					Timestamp: time.Unix(10, 0),
				},
				{
					Pair:      Pair{Base: "B", Quote: "C"},
					Price:     0,
					Bid:       0,
					Ask:       0,
					Volume24h: 10,
					Timestamp: time.Unix(10, 0),
				},
			},
			want:    Tick{},
			wantErr: true,
		},
		{
			name: "ca/cb",
			ticks: []Tick{
				{
					Pair:      Pair{Base: "C", Quote: "A"},
					Price:     10,
					Bid:       5,
					Ask:       15,
					Volume24h: 10,
					Timestamp: time.Unix(10, 0),
				},
				{
					Pair:      Pair{Base: "C", Quote: "B"},
					Price:     20,
					Bid:       10,
					Ask:       30,
					Volume24h: 10,
					Timestamp: time.Unix(10, 0),
				},
			},
			want: Tick{
				Pair:      Pair{Base: "A", Quote: "C"},
				Price:     float64(20) / 10,
				Bid:       float64(10) / 5,
				Ask:       float64(30) / 15,
				Volume24h: 10,
				Timestamp: time.Unix(10, 0),
			},
			wantErr: false,
		},
		{
			name: "ca/cb-divByZero",
			ticks: []Tick{
				{
					Pair:      Pair{Base: "C", Quote: "A"},
					Price:     0,
					Bid:       0,
					Ask:       0,
					Volume24h: 0,
					Timestamp: time.Unix(10, 0),
				},
				{
					Pair:      Pair{Base: "C", Quote: "B"},
					Price:     20,
					Bid:       10,
					Ask:       30,
					Volume24h: 10,
					Timestamp: time.Unix(10, 0),
				},
			},
			want:    Tick{},
			wantErr: true,
		},
		{
			name: "ac/cb",
			ticks: []Tick{
				{
					Pair:      Pair{Base: "A", Quote: "C"},
					Price:     10,
					Bid:       5,
					Ask:       15,
					Volume24h: 10,
					Timestamp: time.Unix(10, 0),
				},
				{
					Pair:      Pair{Base: "C", Quote: "B"},
					Price:     20,
					Bid:       10,
					Ask:       30,
					Volume24h: 10,
					Timestamp: time.Unix(10, 0),
				},
			},
			want: Tick{
				Pair:      Pair{Base: "A", Quote: "C"},
				Price:     float64(20) * 10,
				Bid:       float64(10) * 5,
				Ask:       float64(30) * 15,
				Volume24h: 10,
				Timestamp: time.Unix(10, 0),
			},
			wantErr: false,
		},
		{
			name: "ca/bc",
			ticks: []Tick{
				{
					Pair:      Pair{Base: "C", Quote: "A"},
					Price:     10,
					Bid:       5,
					Ask:       15,
					Volume24h: 10,
					Timestamp: time.Unix(10, 0),
				},
				{
					Pair:      Pair{Base: "B", Quote: "C"},
					Price:     20,
					Bid:       10,
					Ask:       30,
					Volume24h: 10,
					Timestamp: time.Unix(10, 0),
				},
			},
			want: Tick{
				Pair:      Pair{Base: "A", Quote: "C"},
				Price:     float64(1) / 20 / 10,
				Bid:       float64(1) / 10 / 5,
				Ask:       float64(1) / 30 / 15,
				Volume24h: 10,
				Timestamp: time.Unix(10, 0),
			},
			wantErr: false,
		},
		{
			name: "ca/bc-divByZero1",
			ticks: []Tick{
				{
					Pair:      Pair{Base: "C", Quote: "A"},
					Price:     10,
					Bid:       5,
					Ask:       15,
					Volume24h: 10,
					Timestamp: time.Unix(10, 0),
				},
				{
					Pair:      Pair{Base: "B", Quote: "C"},
					Price:     0,
					Bid:       0,
					Ask:       0,
					Volume24h: 0,
					Timestamp: time.Unix(10, 0),
				},
			},
			want:    Tick{},
			wantErr: true,
		},
		{
			name: "ca/bc-divByZero2",
			ticks: []Tick{
				{
					Pair:      Pair{Base: "C", Quote: "A"},
					Price:     0,
					Bid:       0,
					Ask:       0,
					Volume24h: 0,
					Timestamp: time.Unix(10, 0),
				},
				{
					Pair:      Pair{Base: "B", Quote: "C"},
					Price:     20,
					Bid:       10,
					Ask:       30,
					Volume24h: 10,
					Timestamp: time.Unix(10, 0),
				},
			},
			want:    Tick{},
			wantErr: true,
		},
		{
			name: "lowest-timestamp",
			ticks: []Tick{
				{
					Pair:      Pair{Base: "A", Quote: "B"},
					Price:     1,
					Bid:       1,
					Ask:       1,
					Volume24h: 0,
					Timestamp: time.Unix(10, 0),
				},
				{
					Pair:      Pair{Base: "B", Quote: "C"},
					Price:     1,
					Bid:       1,
					Ask:       1,
					Volume24h: 0,
					Timestamp: time.Unix(5, 0),
				},
				{
					Pair:      Pair{Base: "C", Quote: "D"},
					Price:     1,
					Bid:       1,
					Ask:       1,
					Volume24h: 0,
					Timestamp: time.Unix(15, 0),
				},
			},
			want: Tick{
				Pair:      Pair{Base: "A", Quote: "D"},
				Price:     1,
				Bid:       1,
				Ask:       1,
				Volume24h: 0,
				Timestamp: time.Unix(5, 0),
			},
			wantErr: false,
		},
		{
			name: "five-pairs",
			ticks: []Tick{
				{
					Pair:      Pair{Base: "ETH", Quote: "BTC"}, // -> ETH/BTC
					Price:     0.050,
					Bid:       0.040,
					Ask:       0.060,
					Volume24h: 10,
					Timestamp: time.Unix(10, 0),
				},
				{
					Pair:      Pair{Base: "BTC", Quote: "USD"}, // -> ETH/USD
					Price:     10000.000,
					Bid:       9000.000,
					Ask:       11000.000,
					Volume24h: 15,
					Timestamp: time.Unix(9, 0),
				},
				{
					Pair:      Pair{Base: "EUR", Quote: "USD"}, // -> ETH/EUR
					Price:     1.250,
					Bid:       1.200,
					Ask:       1.300,
					Volume24h: 20,
					Timestamp: time.Unix(11, 0),
				},
				{
					Pair:      Pair{Base: "EUR", Quote: "CAD"}, // -> ETH/CAD
					Price:     1.250,
					Bid:       1.200,
					Ask:       1.300,
					Volume24h: 25,
					Timestamp: time.Unix(8, 0),
				},
				{
					Pair:      Pair{Base: "GPB", Quote: "ETH"}, // -> GPB/CAD
					Price:     0.005,
					Bid:       0.004,
					Ask:       0.006,
					Volume24h: 30,
					Timestamp: time.Unix(13, 0),
				},
			},
			want: Tick{
				Pair:      Pair{Base: "GPB", Quote: "CAD"},
				Price:     float64(1) / (((float64(0.050) * 10000.000) / 1.250) * 1.250) / 0.005,
				Bid:       float64(1) / (((float64(0.040) * 9000.000) / 1.200) * 1.200) / 0.004,
				Ask:       float64(1) / (((float64(0.060) * 11000.000) / 1.300) * 1.300) / 0.006,
				Volume24h: 0,
				Timestamp: time.Unix(8, 0),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := crossRate(tt.ticks)

			if err != nil {
				if tt.wantErr {
					assert.Error(t, err)
				}

				return
			}

			assert.InDelta(t, tt.want.Price, got.Price, 0.000000001)
			assert.InDelta(t, tt.want.Bid, got.Bid, 0.000000001)
			assert.InDelta(t, tt.want.Ask, got.Ask, 0.000000001)
			assert.Equal(t, tt.want.Timestamp.Unix(), got.Timestamp.Unix())
			assert.Nil(t, err)
		})
	}
}
