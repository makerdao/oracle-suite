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

package config

import (
	"encoding/json"

	"github.com/makerdao/oracle-suite/internal/query"
	"github.com/makerdao/oracle-suite/pkg/gofer/origins"
)

type apiKeyParams struct {
	APIKey string `json:"apiKey"`
}

func parseParamsToAPIKey(params json.RawMessage) string {
	if params == nil {
		return ""
	}

	var res apiKeyParams
	err := json.Unmarshal(params, &res)
	if err != nil {
		return ""
	}
	return res.APIKey
}

//nolint
func NewHandler(handlerType string, pool query.WorkerPool, params json.RawMessage) origins.Handler {
	switch handlerType {
	case "balancer":
		return &origins.Balancer{Pool: pool}
	case "binance":
		return &origins.Binance{Pool: pool}
	case "bitfinex":
		return &origins.Bitfinex{Pool: pool}
	case "bitstamp":
		return &origins.Bitstamp{Pool: pool}
	case "bitthumb":
		return &origins.BitThump{Pool: pool}
	case "bithumb":
		return &origins.BitThump{Pool: pool}
	case "bittrex":
		return &origins.Bittrex{Pool: pool}
	case "coinbase", "coinbasepro":
		return &origins.CoinbasePro{Pool: pool}
	case "cryptocompare":
		return &origins.CryptoCompare{Pool: pool}
	case "coinmarketcap":
		apiKey := parseParamsToAPIKey(params)
		return &origins.CoinMarketCap{Pool: pool, APIKey: apiKey}
	case "ddex":
		return &origins.Ddex{Pool: pool}
	case "folgory":
		return &origins.Folgory{Pool: pool}
	case "ftx":
		return &origins.Ftx{Pool: pool}
	case "fx":
		return &origins.Fx{Pool: pool}
	case "gateio":
		return &origins.Gateio{Pool: pool}
	case "gemini":
		return &origins.Gemini{Pool: pool}
	case "hitbtc":
		return &origins.Hitbtc{Pool: pool}
	case "huobi":
		return &origins.Huobi{Pool: pool}
	case "kraken":
		return &origins.Kraken{Pool: pool}
	case "kucoin":
		return &origins.Kucoin{Pool: pool}
	case "kyber":
		return &origins.Kyber{Pool: pool}
	case "loopring":
		return &origins.Loopring{Pool: pool}
	case "okex":
		return &origins.Okex{Pool: pool}
	case "operexchangerates":
		apiKey := parseParamsToAPIKey(params)
		return &origins.OpenExchangeRates{Pool: pool, APIKey: apiKey}
	case "poloniex":
		return &origins.Poloniex{Pool: pool}
	case "sushiswap":
		return &origins.Sushiswap{Pool: pool}
	case "uniswap":
		return &origins.Uniswap{Pool: pool}
	case "upbit":
		return &origins.Upbit{Pool: pool}
	}

	return nil
}

func DefaultOriginSet(pool query.WorkerPool) *origins.Set {
	return origins.NewSet(map[string]origins.Handler{
		//"coinmarketcap":     &origins.Balancer{Pool: pool},
		//"operexchangerates": &origins.Balancer{Pool: pool},
		"balancer":      &origins.Balancer{Pool: pool},
		"binance":       &origins.Binance{Pool: pool},
		"bitfinex":      &origins.Bitfinex{Pool: pool},
		"bitstamp":      &origins.Bitstamp{Pool: pool},
		"bitthumb":      &origins.BitThump{Pool: pool},
		"bithumb":       &origins.BitThump{Pool: pool},
		"bittrex":       &origins.Bittrex{Pool: pool},
		"coinbase":      &origins.CoinbasePro{Pool: pool},
		"coinbasepro":   &origins.CoinbasePro{Pool: pool},
		"cryptocompare": &origins.CryptoCompare{Pool: pool},
		"ddex":          &origins.Ddex{Pool: pool},
		"folgory":       &origins.Folgory{Pool: pool},
		"ftx":           &origins.Ftx{Pool: pool},
		"fx":            &origins.Fx{Pool: pool},
		"gateio":        &origins.Gateio{Pool: pool},
		"gemini":        &origins.Gemini{Pool: pool},
		"hitbtc":        &origins.Hitbtc{Pool: pool},
		"huobi":         &origins.Huobi{Pool: pool},
		"kraken":        &origins.Kraken{Pool: pool},
		"kucoin":        &origins.Kucoin{Pool: pool},
		"kyber":         &origins.Kyber{Pool: pool},
		"loopring":      &origins.Loopring{Pool: pool},
		"okex":          &origins.Okex{Pool: pool},
		"poloniex":      &origins.Poloniex{Pool: pool},
		"sushiswap":     &origins.Sushiswap{Pool: pool},
		"uniswap":       &origins.Uniswap{Pool: pool},
		"upbit":         &origins.Upbit{Pool: pool},
	})
}
