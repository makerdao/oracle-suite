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

	"github.com/makerdao/gofer/model"
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
