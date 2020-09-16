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
)

func Test_calcIndirectTick(t *testing.T) {
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
			got, err := calcIndirectTick(tt.ticks)

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
