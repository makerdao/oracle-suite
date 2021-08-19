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
		return origins.NewBaseExchangeHandler(origins.Balancer{
			WorkerPool:        pool,
			ContractAddresses: parsedParams.Contracts,
		}, parsedParams.SymbolAliases), nil
	case "binance":
		return origins.NewBaseExchangeHandler(origins.Binance{WorkerPool: pool}, parsedParams.SymbolAliases), nil
	case "bitfinex":
		return origins.NewBaseExchangeHandler(origins.Bitfinex{WorkerPool: pool}, parsedParams.SymbolAliases), nil
	case "bitstamp":
		return origins.NewBaseExchangeHandler(origins.Bitstamp{WorkerPool: pool}, parsedParams.SymbolAliases), nil
	case "bitthumb", "bithumb":
		return origins.NewBaseExchangeHandler(origins.BitThump{WorkerPool: pool}, parsedParams.SymbolAliases), nil
	case "bittrex":
		return origins.NewBaseExchangeHandler(origins.Bittrex{WorkerPool: pool}, parsedParams.SymbolAliases), nil
	case "coinbase", "coinbasepro":
		return origins.NewBaseExchangeHandler(origins.CoinbasePro{WorkerPool: pool}, parsedParams.SymbolAliases), nil
	case "cryptocompare":
		return origins.NewBaseExchangeHandler(origins.CryptoCompare{WorkerPool: pool}, parsedParams.SymbolAliases), nil
	case "coinmarketcap":
		return origins.NewBaseExchangeHandler(
			origins.CoinMarketCap{WorkerPool: pool, APIKey: parsedParams.APIKey},
			parsedParams.SymbolAliases,
		), nil
	case "ddex":
		return origins.NewBaseExchangeHandler(origins.Ddex{WorkerPool: pool}, parsedParams.SymbolAliases), nil
	case "folgory":
		return origins.NewBaseExchangeHandler(origins.Folgory{WorkerPool: pool}, parsedParams.SymbolAliases), nil
	case "ftx":
		return origins.NewBaseExchangeHandler(origins.Ftx{WorkerPool: pool}, parsedParams.SymbolAliases), nil
	case "fx":
		return origins.NewBaseExchangeHandler(
			origins.Fx{WorkerPool: pool, APIKey: parsedParams.APIKey},
			parsedParams.SymbolAliases,
		), nil
	case "gateio":
		return origins.NewBaseExchangeHandler(origins.Gateio{WorkerPool: pool}, parsedParams.SymbolAliases), nil
	case "gemini":
		return origins.NewBaseExchangeHandler(origins.Gemini{WorkerPool: pool}, parsedParams.SymbolAliases), nil
	case "hitbtc":
		return origins.NewBaseExchangeHandler(origins.Hitbtc{WorkerPool: pool}, parsedParams.SymbolAliases), nil
	case "huobi":
		return origins.NewBaseExchangeHandler(origins.Huobi{WorkerPool: pool}, parsedParams.SymbolAliases), nil
	case "kraken":
		return origins.NewBaseExchangeHandler(origins.Kraken{WorkerPool: pool}, parsedParams.SymbolAliases), nil
	case "kucoin":
		return origins.NewBaseExchangeHandler(origins.Kucoin{WorkerPool: pool}, parsedParams.SymbolAliases), nil
	case "kyber":
		return origins.NewBaseExchangeHandler(origins.Kyber{WorkerPool: pool}, parsedParams.SymbolAliases), nil
	case "loopring":
		return origins.NewBaseExchangeHandler(origins.Loopring{WorkerPool: pool}, parsedParams.SymbolAliases), nil
	case "okex":
		return origins.NewBaseExchangeHandler(origins.Okex{WorkerPool: pool}, parsedParams.SymbolAliases), nil
	case "openexchangerates":
		return origins.NewBaseExchangeHandler(
			origins.OpenExchangeRates{WorkerPool: pool, APIKey: parsedParams.APIKey},
			parsedParams.SymbolAliases,
		), nil
	case "poloniex":
		return origins.NewBaseExchangeHandler(origins.Poloniex{WorkerPool: pool}, parsedParams.SymbolAliases), nil
	case "sushiswap":
		return origins.NewBaseExchangeHandler(origins.Sushiswap{
			WorkerPool:        pool,
			ContractAddresses: parsedParams.Contracts,
		}, parsedParams.SymbolAliases), nil
	case "uniswap", "uniswapV2":
		return origins.NewBaseExchangeHandler(origins.Uniswap{
			WorkerPool:        pool,
			ContractAddresses: parsedParams.Contracts,
		}, parsedParams.SymbolAliases), nil
	case "uniswapV3":
		return origins.NewBaseExchangeHandler(origins.UniswapV3{
			WorkerPool:        pool,
			ContractAddresses: parsedParams.Contracts,
		}, parsedParams.SymbolAliases), nil
	case "upbit":
		return origins.NewBaseExchangeHandler(origins.Upbit{WorkerPool: pool}, parsedParams.SymbolAliases), nil
	}

	return nil, origins.ErrUnknownOrigin
}
