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

package gofer

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
	case "uniswapV2":
		return &origins.Uniswap{Pool: pool}, nil
	case "uniswapV3":
		return &origins.UniswapV3{Pool: pool}, nil
	case "upbit":
		return &origins.Upbit{Pool: pool}, nil
	}

	return nil, origins.ErrUnknownOrigin
}
