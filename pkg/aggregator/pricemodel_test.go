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

package aggregator

import (
	"testing"

	"github.com/makerdao/gofer/pkg/model"
)

func TestPriceModel_String(t *testing.T) {
	tests := []struct {
		name string
		pm   PriceModel
		want string
	}{{
		"ETH/USD",
		PriceModel{
			Method:     "median",
			MinSources: 1,
			Sources: []PriceRefPath{
				{PriceRef{Origin: "conibase", Pair: Pair{*model.NewPair("ETH", "USD")}}},
				{PriceRef{Origin: "kraken", Pair: Pair{*model.NewPair("ETH", "USD")}}},
			},
		},
		"(median:1)[[ETH/USD@conibase][ETH/USD@kraken]]",
	}, {
		"WBTC/USD",
		PriceModel{
			Method:     "median",
			MinSources: 2,
			Sources: []PriceRefPath{
				{PriceRef{Origin: "coinbase", Pair: Pair{*model.NewPair("WBTC", "USD")}}},
				{
					PriceRef{Origin: "kraken", Pair: Pair{*model.NewPair("WBTC", "ETH")}},
					PriceRef{Origin: ".", Pair: Pair{*model.NewPair("ETH", "USD")}},
				},
			},
		},
		"(median:2)[[WBTC/USD@coinbase][WBTC/ETH@kraken,ETH/USD]]",
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pm.String(); got != tt.want {
				t.Errorf("String() = %s, want %s", got, tt.want)
			}
		})
	}
}

func Test_resolvePath(t *testing.T) {
	tests := []struct {
		name    string
		pas     []*model.PriceAggregate
		want    *model.PriceAggregate
		wantErr bool
	}{
		{
			name:    "resolvePath()=>MKR/USD",
			pas:     []*model.PriceAggregate{},
			wantErr: true,
		}, {
			name: "resolvePath(MKR/USD)=>MKR/USD",
			pas: []*model.PriceAggregate{
				newTestPricePointAggregate(0, "exchange1", "MKR", "USD", 123, 1),
			},
			want: newTestPricePointAggregate(0, "trade", "MKR", "USD", 123, 1),
		}, {
			name: "resolvePath(MKR/ETH,USD/ETH)=>MKR/USD",
			pas: []*model.PriceAggregate{
				newTestPricePointAggregate(0, "exchange1", "MKR", "ETH", 10, 1),
				newTestPricePointAggregate(0, "exchange2", "USD", "ETH", 20, 1),
			},
			want: newTestPricePointAggregate(0, "trade", "MKR", "USD", 0.5, 1),
		}, {
			name: "resolvePath(ETH/MKR,USD/ETH)=>MKR/USD",
			pas: []*model.PriceAggregate{
				newTestPricePointAggregate(0, "exchange1", "ETH", "MKR", 10, 1),
				newTestPricePointAggregate(0, "exchange2", "USD", "ETH", 20, 1),
			},
			want: newTestPricePointAggregate(0, "trade", "MKR", "USD", 0.005, 1),
		}, {
			name: "resolvePath(MKR/ETH,ETH/USD)=>MKR/USD",
			pas: []*model.PriceAggregate{
				newTestPricePointAggregate(0, "exchange1", "MKR", "ETH", 10, 1),
				newTestPricePointAggregate(0, "exchange2", "ETH", "USD", 20, 1),
			},
			want: newTestPricePointAggregate(0, "trade", "MKR", "USD", 200, 1),
		}, {
			name: "resolvePath(ETH/MKR,ETH/USD)=>MKR/USD",
			pas: []*model.PriceAggregate{
				newTestPricePointAggregate(0, "exchange1", "ETH", "MKR", 5, 1),
				newTestPricePointAggregate(0, "exchange2", "ETH", "USD", 20, 1),
			},
			want: newTestPricePointAggregate(0, "trade", "MKR", "USD", 4, 1),
		}, {
			name: "resolvePath(ETH/MKR,ETH/BTC,BTC/USD)=>MKR/USD",
			pas: []*model.PriceAggregate{
				newTestPricePointAggregate(0, "exchange1", "ETH", "MKR", 5, 1),
				newTestPricePointAggregate(0, "exchange2", "ETH", "BTC", 20, 1),
				newTestPricePointAggregate(0, "exchange2", "BTC", "USD", 3, 1),
			},
			want: newTestPricePointAggregate(0, "trade", "MKR", "USD", 12, 1),
		}, {
			name: "convert(ETH/MKR,BTC/USD)=>error",
			pas: []*model.PriceAggregate{
				newTestPricePointAggregate(0, "exchange1", "ETH", "MKR", 10, 1),
				newTestPricePointAggregate(0, "exchange2", "BTC", "USD", 20, 1),
			},
			wantErr: true,
		}, {
			name: "convert(ETH/MKR,ETH/BTC,USDT/USD)=>error",
			pas: []*model.PriceAggregate{
				newTestPricePointAggregate(0, "exchange1", "ETH", "MKR", 5, 1),
				newTestPricePointAggregate(0, "exchange2", "ETH", "BTC", 20, 1),
				newTestPricePointAggregate(0, "exchange2", "USDT", "USD", 3, 1),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolvePath(tt.pas)
			if (err != nil) != tt.wantErr {
				t.Errorf("resolvePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				// add failing tests here
				return
			}
			if got.Pair.String() != tt.want.Pair.String() {
				t.Errorf("resolvePath() got = %s, want %s", got, tt.want)
				return
			}
			if got.Price != tt.want.Price {
				t.Errorf("resolvePath() got = %f, want %f", got.Price, tt.want.Price)
			}
		})
	}
}

func Test_ResolveRef(t *testing.T) {
	tests := []struct {
		name    string
		pmm     PriceModelMap
		cache   PriceCache
		pair    Pair
		want    *model.PriceAggregate
		wantErr bool
	}{
		{
			name: "1 successful direct source, no limit",
			pmm: PriceModelMap{
				pair("A", "B"): PriceModel{
					Method:     "median",
					MinSources: 0,
					Sources: []PriceRefPath{
						{PriceRef{Origin: "E-1", Pair: pair("A", "B")}},
					},
				},
			},
			cache: PriceCache{
				{pair: pair("A", "B"), exchangeName: "E-1"}: newTestPriceAggregate("", "A", "B", 12.0),
			},
			pair:    pair("A", "B"),
			want:    newTestPriceAggregate("median", "A", "B", 12.0, newTestPriceAggregate("", "A", "B", 12.0)),
			wantErr: false,
		},
		{
			name: "1 failed direct source, no limit",
			pmm: PriceModelMap{
				pair("A", "B"): PriceModel{
					Method:     "median",
					MinSources: 0,
					Sources: []PriceRefPath{
						{PriceRef{Origin: "E-1", Pair: pair("A", "B")}},
					},
				},
			},
			cache:   PriceCache{},
			pair:    pair("A", "B"),
			wantErr: true,
		},
		{
			name: "2 successful direct sources, limit 2",
			pmm: PriceModelMap{
				pair("A", "B"): PriceModel{
					Method:     "median",
					MinSources: 2,
					Sources: []PriceRefPath{
						{PriceRef{Origin: "E-1", Pair: pair("A", "B")}},
						{PriceRef{Origin: "E-2", Pair: pair("A", "B")}},
					},
				},
			},
			cache: PriceCache{
				{pair: pair("A", "B"), exchangeName: "E-1"}: newTestPriceAggregate("", "A", "B", 12.0),
				{pair: pair("A", "B"), exchangeName: "E-2"}: newTestPriceAggregate("", "A", "B", 14.0),
			},
			pair: pair("A", "B"),
			want: newTestPriceAggregate(
				"median",
				"A",
				"B",
				13.0,
				newTestPriceAggregate("", "A", "B", 12.0),
				newTestPriceAggregate("", "A", "B", 14.0),
			),
			wantErr: false,
		},
		{
			name: "2 direct sources, 1 successful, limit 2",
			pmm: PriceModelMap{
				pair("A", "B"): PriceModel{
					Method:     "median",
					MinSources: 2,
					Sources: []PriceRefPath{
						{PriceRef{Origin: "E-1", Pair: pair("A", "B")}},
						{PriceRef{Origin: "E-2", Pair: pair("A", "B")}},
					},
				},
			},
			cache: PriceCache{
				{pair: pair("A", "B"), exchangeName: "E-1"}: newTestPriceAggregate("", "A", "B", 12.0),
			},
			pair:    pair("A", "B"),
			wantErr: true,
		},
		{
			name: "1 successful direct + 1 failing indirect source, limit 1",
			pmm: PriceModelMap{
				pair("A", "C"): PriceModel{
					Method:     "median",
					MinSources: 1,
					Sources: []PriceRefPath{
						{PriceRef{Origin: "E-1", Pair: pair("A", "B")}, PriceRef{Origin: ".", Pair: pair("B", "C")}},
						{PriceRef{Origin: "E-1", Pair: pair("A", "C")}},
					},
				},
				pair("B", "C"): PriceModel{
					Method:     "median",
					MinSources: 2,
					Sources: []PriceRefPath{
						{PriceRef{Origin: "E-1", Pair: pair("B", "C")}},
						{PriceRef{Origin: "E-2", Pair: pair("B", "C")}},
					},
				},
			},
			cache: PriceCache{
				{pair: pair("A", "B"), exchangeName: "E-1"}: newTestPriceAggregate("", "A", "B", 12.0),
				{pair: pair("A", "C"), exchangeName: "E-1"}: newTestPriceAggregate("", "A", "C", 14.0),
				{pair: pair("B", "C"), exchangeName: "E-1"}: newTestPriceAggregate("", "B", "C", 16.0),
			},
			pair:    pair("A", "C"),
			want:    newTestPriceAggregate("median", "A", "C", 14.0, newTestPriceAggregate("", "A", "C", 14.0)),
			wantErr: false,
		},
		{
			name: "1 failing direct + 1 successful indirect source, limit 1",
			pmm: PriceModelMap{
				pair("A", "C"): PriceModel{
					Method:     "median",
					MinSources: 1,
					Sources: []PriceRefPath{
						{PriceRef{Origin: "E-1", Pair: pair("A", "B")}, PriceRef{Origin: ".", Pair: pair("B", "C")}},
						{PriceRef{Origin: "E-1", Pair: pair("A", "C")}},
					},
				},
				pair("B", "C"): PriceModel{
					Method:     "median",
					MinSources: 2,
					Sources: []PriceRefPath{
						{PriceRef{Origin: "E-1", Pair: pair("B", "C")}},
						{PriceRef{Origin: "E-2", Pair: pair("B", "C")}},
					},
				},
			},
			cache: PriceCache{
				{pair: pair("A", "B"), exchangeName: "E-1"}: newTestPriceAggregate("", "A", "B", 12.0),
				{pair: pair("B", "C"), exchangeName: "E-1"}: newTestPriceAggregate("", "B", "C", 16.0),
				{pair: pair("B", "C"), exchangeName: "E-2"}: newTestPriceAggregate("", "B", "C", 18.0),
			},
			pair: pair("A", "C"),
			want: newTestPriceAggregate(
				"median",
				"A",
				"C",
				12.0*17.0,
				newTestPriceAggregate(
					"trade",
					"A",
					"C",
					12.0*17.0,
					newTestPriceAggregate("", "A", "B", 12.0),
					newTestPriceAggregate(
						"median",
						"B",
						"C",
						17.0,
						newTestPriceAggregate("", "B", "C", 16.0),
						newTestPriceAggregate("", "B", "C", 18.0),
					),
				),
			),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.pmm.ResolveRef(tt.cache, PriceRef{Origin: ".", Pair: tt.pair})
			if (err != nil) != tt.wantErr {
				t.Errorf("PriceModelMap.ResolveRef() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if (tt.want == nil && got != tt.want) || got.String() != tt.want.String() {
				t.Errorf("PriceModelMap.ResolveRef() got = %s, want %s", got, tt.want)
				return
			}
		})
	}
}

func pair(base string, quote string) Pair {
	return Pair{Pair: model.Pair{Base: base, Quote: quote}}
}
