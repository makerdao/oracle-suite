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
