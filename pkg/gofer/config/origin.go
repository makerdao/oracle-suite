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
	"fmt"

	"github.com/makerdao/oracle-suite/internal/query"
	"github.com/makerdao/oracle-suite/pkg/gofer/origins"
)

type apiKeyParams struct {
	APIKey string `json:"apiKey"`
}

func parseParamsToAPIKey(params json.RawMessage) (string, error) {
	if params == nil {
		return "", fmt.Errorf("invalid origin parameters")
	}

	var res apiKeyParams
	err := json.Unmarshal(params, &res)
	if err != nil {
		return "", fmt.Errorf("failed to marshal origin parameters: %w", err)
	}
	return res.APIKey, nil
}

//nolint
func NewHandler(handlerType string, pool query.WorkerPool, params json.RawMessage) (origins.Handler, error) {
	switch handlerType {
	case "balancer":
		return &origins.Balancer{Pool: pool}, nil
	case "binance":
		return &origins.Binance{Pool: pool}, nil
	case "bitfinex":
		return &origins.Bitfinex{Pool: pool}, nil
	case "bitstamp":
		return &origins.Bitstamp{Pool: pool}, nil
	case "bitthumb":
		return &origins.BitThump{Pool: pool}, nil
	case "bithumb":
		return &origins.BitThump{Pool: pool}, nil
	case "bittrex":
		return &origins.Bittrex{Pool: pool}, nil
	case "coinbase", "coinbasepro":
		return &origins.CoinbasePro{Pool: pool}, nil
	case "cryptocompare":
		return &origins.CryptoCompare{Pool: pool}, nil
	case "coinmarketcap":
		apiKey, err := parseParamsToAPIKey(params)
		if err != nil {
			return nil, err
		}
		return &origins.CoinMarketCap{Pool: pool, APIKey: apiKey}, nil
	case "ddex":
		return &origins.Ddex{Pool: pool}, nil
	case "folgory":
		return &origins.Folgory{Pool: pool}, nil
	case "ftx":
		return &origins.Ftx{Pool: pool}, nil
	case "fx":
		apiKey, err := parseParamsToAPIKey(params)
		if err != nil {
			return nil, err
		}
		return &origins.Fx{Pool: pool, APIKey: apiKey}, nil
	case "gateio":
		return &origins.Gateio{Pool: pool}, nil
	case "gemini":
		return &origins.Gemini{Pool: pool}, nil
	case "hitbtc":
		return &origins.Hitbtc{Pool: pool}, nil
	case "huobi":
		return &origins.Huobi{Pool: pool}, nil
	case "kraken":
		return &origins.Kraken{Pool: pool}, nil
	case "kucoin":
		return &origins.Kucoin{Pool: pool}, nil
	case "kyber":
		return &origins.Kyber{Pool: pool}, nil
	case "loopring":
		return &origins.Loopring{Pool: pool}, nil
	case "okex":
		return &origins.Okex{Pool: pool}, nil
	case "openexchangerates":
		apiKey, err := parseParamsToAPIKey(params)
		if err != nil {
			return nil, err
		}
		return &origins.OpenExchangeRates{Pool: pool, APIKey: apiKey}, nil
	case "poloniex":
		return &origins.Poloniex{Pool: pool}, nil
	case "sushiswap":
		return &origins.Sushiswap{Pool: pool}, nil
	case "uniswap":
		return &origins.Uniswap{Pool: pool}, nil
	case "upbit":
		return &origins.Upbit{Pool: pool}, nil
	}

	return nil, origins.ErrUnknownOrigin
}

func DefaultOriginSet(pool query.WorkerPool) *origins.Set {
	return origins.NewSet(map[string]origins.Handler{
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
