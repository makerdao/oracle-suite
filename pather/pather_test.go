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

package pather

import (
	"github.com/stretchr/testify/assert"
	. "makerdao/gofer/model"
	"testing"
)

func TestPricePathETHBTC(t *testing.T) {
	ppath := NewPricePaths(
		NewPair("ETH", "BTC"),
		[]*Pair{
			NewPair("ETH", "BTC"),
		},
	)
	ppps := []*PotentialPricePoint{
		{
			Pair: NewPair("ETH", "BTC"),
			Exchange: &Exchange{
				Name: "exchange-a",
			},
		},
		{
			Pair: NewPair("BAT", "BTC"),
			Exchange: &Exchange{
				Name: "exchange-a",
			},
		},
	}
	pppsFail := []*PotentialPricePoint{
		{
			Pair: NewPair("BTC", "USD"),
			Exchange: &Exchange{
				Name: "exchange-a",
			},
		},
	}

	var pppRes []*PotentialPricePoint
	pppRes = FilterPotentialPricePoints(ppath, ppps)
	assert.NotNil(t, pppRes)
	assert.Equal(t, pppRes, []*PotentialPricePoint{
		{
			Pair: NewPair("ETH", "BTC"),
			Exchange: &Exchange{
				Name: "exchange-a",
			},
		},
	})

	pppRes = FilterPotentialPricePoints(ppath, pppsFail)
	assert.Nil(t, pppRes)
}

func TestPricePathBATUSD(t *testing.T) {
	ppath := NewPricePaths(
		NewPair("BAT", "USD"),
		[]*Pair{
			NewPair("BAT", "BTC"),
			NewPair("BTC", "USD"),
		},
		[]*Pair{
			NewPair("BAT", "KRW"),
			NewPair("KRW", "USD"),
		},
	)

	ppps_BAT_BTC := []*PotentialPricePoint{
		{
			Pair: NewPair("BTC", "USD"),
			Exchange: &Exchange{
				Name: "exchange-a",
			},
		},
		{
			Pair: NewPair("BAT", "BTC"),
			Exchange: &Exchange{
				Name: "exchange-a",
			},
		},
		{
			Pair: NewPair("BAT", "BTC"),
			Exchange: &Exchange{
				Name: "exchange-b",
			},
		},
		{
			Pair: NewPair("BAT", "KRW"),
			Exchange: &Exchange{
				Name: "exchange-a",
			},
		},
		{
			Pair: NewPair("BAT", "KRW"),
			Exchange: &Exchange{
				Name: "fx",
			},
		},
		{
			Pair: NewPair("BAT", "ETH"),
			Exchange: &Exchange{
				Name: "fx",
			},
		},
		{
			Pair: NewPair("BAT", "ETH"),
			Exchange: &Exchange{
				Name: "fx",
			},
		},
	}
	ppps_BAT_BTC_KRW := append(
		ppps_BAT_BTC,
		&PotentialPricePoint{
			Pair: NewPair("KRW", "USD"),
			Exchange: &Exchange{
				Name: "fx",
			},
		},
	)
	pppsFail := []*PotentialPricePoint{
		{
			Pair: NewPair("KRW", "USD"),
			Exchange: &Exchange{
				Name: "exchange-a",
			},
		},
		{
			Pair: NewPair("BAT", "BTC"),
			Exchange: &Exchange{
				Name: "exchange-a",
			},
		},
		{
			Pair: NewPair("ETH", "USD"),
			Exchange: &Exchange{
				Name: "exchange-b",
			},
		},
	}

	var pppRes []*PotentialPricePoint
	pppRes = FilterPotentialPricePoints(ppath, ppps_BAT_BTC)
	assert.NotNil(t, pppRes)
	assert.ElementsMatch(t, []*PotentialPricePoint{
		{
			Pair: NewPair("BAT", "BTC"),
			Exchange: &Exchange{
				Name: "exchange-a",
			},
		},
		{
			Pair: NewPair("BTC", "USD"),
			Exchange: &Exchange{
				Name: "exchange-a",
			},
		},
		{
			Pair: NewPair("BAT", "BTC"),
			Exchange: &Exchange{
				Name: "exchange-b",
			},
		},
	}, pppRes)

	pppRes = FilterPotentialPricePoints(ppath, ppps_BAT_BTC_KRW)
	assert.NotNil(t, pppRes)
	assert.ElementsMatch(t, []*PotentialPricePoint{
		{
			Pair: NewPair("BAT", "BTC"),
			Exchange: &Exchange{
				Name: "exchange-a",
			},
		},
		{
			Pair: NewPair("BTC", "USD"),
			Exchange: &Exchange{
				Name: "exchange-a",
			},
		},
		{
			Pair: NewPair("BAT", "BTC"),
			Exchange: &Exchange{
				Name: "exchange-b",
			},
		},
		{
			Pair: NewPair("KRW", "USD"),
			Exchange: &Exchange{
				Name: "fx",
			},
		},
		{
			Pair: NewPair("BAT", "KRW"),
			Exchange: &Exchange{
				Name: "exchange-a",
			},
		},
		{
			Pair: NewPair("BAT", "KRW"),
			Exchange: &Exchange{
				Name: "fx",
			},
		},
	}, pppRes)

	pppRes = FilterPotentialPricePoints(ppath, pppsFail)
	assert.Nil(t, pppRes)
}
