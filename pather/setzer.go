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
	"makerdao/gofer/model"
)

type Setzer struct{}

// Create a new Pather based on setzer path model
func NewSetzer() *Setzer {
	return &Setzer{}
}

// Pairs returns a list of Pairs that are tradeable based on setzer
func (sppf *Setzer) Pairs() []*model.Pair {
	return []*model.Pair{
		model.NewPair("BAT", "USD"),
		model.NewPair("BNB", "USD"),
		model.NewPair("BTC", "USD"),
		model.NewPair("DGD", "USD"),
		model.NewPair("DGX", "USD"),
		model.NewPair("ETH", "BTC"),
		model.NewPair("ETH", "USD"),
		model.NewPair("GNT", "USD"),
		model.NewPair("KNC", "USD"),
		model.NewPair("LINK", "USD"),
		model.NewPair("MANA", "USD"),
		model.NewPair("MKR", "USD"),
		model.NewPair("OMG", "USD"),
		model.NewPair("POLY", "USD"),
		model.NewPair("REP", "USD"),
		model.NewPair("SNT", "USD"),
		model.NewPair("USDC", "USD"),
		model.NewPair("USDT", "USD"),
		model.NewPair("WBTC", "USD"),
		model.NewPair("ZRX", "USD"),
	}
}

// Path returns PricePaths describing how to trade between two assets,
// emulating the setzer price path
func (sppf *Setzer) Path(target *model.Pair) *model.PricePaths {
	switch target.String() {
	case "BAT/USD":
		return model.NewPricePaths(
			model.NewPair("BAT", "USD"),
			[]*model.Pair{
				model.NewPair("BAT", "BTC"),
				model.NewPair("BTC", "USD"),
			},
			[]*model.Pair{
				model.NewPair("BAT", "KRW"),
				model.NewPair("KRW", "USD"),
			},
		)
	case "BNB/USD":
		return model.NewPricePaths(
			model.NewPair("BNB", "USD"),
			[]*model.Pair{
				model.NewPair("BNB", "USDC"),
				model.NewPair("USDC", "USD"),
			},
			[]*model.Pair{
				model.NewPair("BNB", "USDT"),
				model.NewPair("USDT", "USD"),
			},
		)
	case "BTC/USD":
		return model.NewPricePaths(
			model.NewPair("BTC", "USD"),
			[]*model.Pair{
				model.NewPair("BTC", "USD"),
			},
		)
	case "DGD/USD":
		return model.NewPricePaths(
			model.NewPair("DGD", "USD"),
			[]*model.Pair{
				model.NewPair("DGD", "BTC"),
				model.NewPair("BTC", "USD"),
			},
		)
	case "DGX/USD":
		return model.NewPricePaths(
			model.NewPair("DGX", "USD"),
			[]*model.Pair{
				model.NewPair("DGX", "USDT"),
				model.NewPair("USDT", "USD"),
			},
			[]*model.Pair{
				model.NewPair("DGX", "ETH"),
				model.NewPair("ETH", "USD"),
			},
		)
	case "ETH/BTC":
		return model.NewPricePaths(
			model.NewPair("ETH", "BTC"),
			[]*model.Pair{
				model.NewPair("ETH", "BTC"),
			},
		)
	case "ETH/USD":
		return model.NewPricePaths(
			model.NewPair("ETH", "USD"),
			[]*model.Pair{
				model.NewPair("ETH", "USD"),
			},
			[]*model.Pair{
				model.NewPair("ETH", "BTC"),
				model.NewPair("BTC", "USD"),
			},
			[]*model.Pair{
				model.NewPair("ETH", "USDT"),
				model.NewPair("USDT", "USD"),
			},
		)
	case "GNT/USD":
		return model.NewPricePaths(
			model.NewPair("GNT", "USD"),
			[]*model.Pair{
				model.NewPair("GNT", "BTC"),
				model.NewPair("BTC", "USD"),
			},
			[]*model.Pair{
				model.NewPair("GNT", "USDC"),
				model.NewPair("USDC", "USD"),
			},
			[]*model.Pair{
				model.NewPair("GNT", "KRW"),
				model.NewPair("KRW", "USD"),
			},
		)
	case "KNC/USD":
		return model.NewPricePaths(
			model.NewPair("KNC", "USD"),
			[]*model.Pair{
				model.NewPair("KNC", "USD"),
			},
			[]*model.Pair{
				model.NewPair("KNC", "BTC"),
				model.NewPair("BTC", "USD"),
			},
			[]*model.Pair{
				model.NewPair("KNC", "ETH"),
				model.NewPair("ETH", "USD"),
			},
			[]*model.Pair{
				model.NewPair("KNC", "KRW"),
				model.NewPair("KRW", "USD"),
			},
		)
	case "LINK/USD":
		return model.NewPricePaths(
			model.NewPair("LINK", "USD"),
			[]*model.Pair{
				model.NewPair("LINK", "USD"),
			},
			[]*model.Pair{
				model.NewPair("LINK", "USDT"),
				model.NewPair("USDT", "USD"),
			},
		)
	case "MANA/USD":
		return model.NewPricePaths(
			model.NewPair("MANA", "USD"),
			[]*model.Pair{
				model.NewPair("MANA", "BTC"),
				model.NewPair("BTC", "USD"),
			},
			[]*model.Pair{
				model.NewPair("MANA", "KRW"),
				model.NewPair("KRW", "USD"),
			},
		)
	case "MKR/USD":
		return model.NewPricePaths(
			model.NewPair("MKR", "USD"),
			[]*model.Pair{
				model.NewPair("MKR", "BTC"),
				model.NewPair("BTC", "USD"),
			},
			[]*model.Pair{
				model.NewPair("MKR", "ETH"),
				model.NewPair("ETH", "USD"),
			},
		)
	case "OMG/USD":
		return model.NewPricePaths(
			model.NewPair("OMG", "USD"),
			[]*model.Pair{
				model.NewPair("OMG", "BTC"),
				model.NewPair("BTC", "USD"),
			},
			[]*model.Pair{
				model.NewPair("OMG", "USDT"),
				model.NewPair("USDT", "USD"),
			},
			[]*model.Pair{
				model.NewPair("OMG", "KRW"),
				model.NewPair("KRW", "USD"),
			},
		)
	case "POLY/USD":
		return model.NewPricePaths(
			model.NewPair("POLY", "USD"),
			[]*model.Pair{
				model.NewPair("POLY", "USD"),
			},
			[]*model.Pair{
				model.NewPair("POLY", "BTC"),
				model.NewPair("BTC", "USD"),
			},
		)
	case "REP/USD":
		return model.NewPricePaths(
			model.NewPair("REP", "USD"),
			[]*model.Pair{
				model.NewPair("REP", "USD"),
			},
			[]*model.Pair{
				model.NewPair("REP", "BTC"),
				model.NewPair("BTC", "USD"),
			},
			[]*model.Pair{
				model.NewPair("REP", "EUR"),
				model.NewPair("EUR", "USD"),
			},
			[]*model.Pair{
				model.NewPair("REP", "KRW"),
				model.NewPair("KRW", "USD"),
			},
		)
	case "SNT/USD":
		return model.NewPricePaths(
			model.NewPair("SNT", "USD"),
			[]*model.Pair{
				model.NewPair("SNT", "BTC"),
				model.NewPair("BTC", "USD"),
			},
			[]*model.Pair{
				model.NewPair("SNT", "USDT"),
				model.NewPair("USDT", "USD"),
			},
			[]*model.Pair{
				model.NewPair("SNT", "KRW"),
				model.NewPair("KRW", "USD"),
			},
		)
	case "USDC/USD":
		return model.NewPricePaths(
			model.NewPair("USDC", "USD"),
			[]*model.Pair{
				model.NewPair("USDC", "USD"),
			},
			// TODO: This path should be added but (A/C = B/C / B/A) is not yet supported
			//[]*model.Pair{
			//	model.NewPair("BTC", "USDC"),
			//	model.NewPair("BTC", "USD"),
			//},
		)
	case "USDT/USD":
		return model.NewPricePaths(
			model.NewPair("USDT", "USD"),
			[]*model.Pair{
				model.NewPair("USDT", "USD"),
			},
			// TODO: This path should be added but (A/C = B/C / B/A) is not yet supported
			//[]*model.Pair{
			//	model.NewPair("BTC", "USDT"),
			//	model.NewPair("BTC", "USD"),
			//},
		)
	case "WBTC/USD":
		return model.NewPricePaths(
			model.NewPair("WBTC", "USD"),
			[]*model.Pair{
				model.NewPair("WBTC", "ETH"),
				model.NewPair("ETH", "USD"),
			},
			[]*model.Pair{
				model.NewPair("WBTC", "USDT"),
				model.NewPair("USDT", "USD"),
			},
		)
	case "ZRX/USD":
		return model.NewPricePaths(
			model.NewPair("ZRX", "USD"),
			[]*model.Pair{
				model.NewPair("ZRX", "USD"),
			},
			[]*model.Pair{
				model.NewPair("ZRX", "BTC"),
				model.NewPair("BTC", "USD"),
			},
			[]*model.Pair{
				model.NewPair("ZRX", "USDT"),
				model.NewPair("USDT", "USD"),
			},
			[]*model.Pair{
				model.NewPair("ZRX", "KRW"),
				model.NewPair("KRW", "USD"),
			},
		)
	}

	return nil
}
