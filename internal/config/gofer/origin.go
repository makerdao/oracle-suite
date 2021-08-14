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

type originParams struct {
	APIKey        string                    `json:"apiKey"`
	Contracts     origins.ContractAddresses `json:"contracts"`
	SymbolAliases origins.SymbolAliases     `json:"symbolAliases"`
}

func parseOriginParams(params json.RawMessage) (*originParams, error) {
	if params == nil {
		return nil, fmt.Errorf("invalid origin parameters")
	}

	var res originParams
	err := json.Unmarshal(params, &res)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal origin parameters: %w", err)
	}
	return &res, nil
}

//nolint
func NewHandler(handlerType string, pool query.WorkerPool, params json.RawMessage) (origins.Handler, error) {
	parsedParams, err := parseOriginParams(params)
	if err != nil {
		return nil, err
	}

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
		return &origins.CoinMarketCap{Pool: pool, APIKey: parsedParams.APIKey}, nil
	case "ddex":
		return &origins.Ddex{Pool: pool}, nil
	case "folgory":
		return &origins.Folgory{Pool: pool}, nil
	case "ftx":
		return &origins.Ftx{Pool: pool}, nil
	case "fx":
		return &origins.Fx{Pool: pool, APIKey: parsedParams.APIKey}, nil
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
		return &origins.OpenExchangeRates{Pool: pool, APIKey: parsedParams.APIKey}, nil
	case "poloniex":
		return &origins.Poloniex{Pool: pool}, nil
	case "sushiswap":
		return &origins.Sushiswap{Pool: pool}, nil
	case "uniswap", "uniswapV2":
		return &origins.Uniswap{
			Pool:              pool,
			ContractAddresses: parsedParams.Contracts,
			SymbolAliases:     parsedParams.SymbolAliases,
		}, nil
	case "uniswapV3":
		return &origins.UniswapV3{
			Pool:              pool,
			ContractAddresses: parsedParams.Contracts,
			SymbolAliases:     parsedParams.SymbolAliases,
		}, nil
	case "upbit":
		return &origins.Upbit{Pool: pool}, nil
	}

	return nil, origins.ErrUnknownOrigin
}
