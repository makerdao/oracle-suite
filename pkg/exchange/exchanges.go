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

package exchange

import (
	"fmt"

	"github.com/makerdao/gofer/internal/query"
	"github.com/makerdao/gofer/pkg/model"
)

// Handler is interface that all Exchange API handlers should implement
type Handler interface {
	// Call should implement making API request to exchange URL and collecting/parsing exchange data
	Call(ppps []*model.PotentialPricePoint) ([]*model.PricePoint, error)
}

type Set struct {
	list map[string]Handler
}

func NewSet(list map[string]Handler) *Set {
	return &Set{list: list}
}

func DefaultSet() *Set {
	const defaultWorkerCount = 5
	httpWorkerPool := query.NewHTTPWorkerPool(defaultWorkerCount)

	return NewSet(map[string]Handler{
		"binance":       &Binance{Pool: httpWorkerPool},
		"bitfinex":      &Bitfinex{Pool: httpWorkerPool},
		"bitstamp":      &Bitstamp{Pool: httpWorkerPool},
		"bittrex":       &BitTrex{Pool: httpWorkerPool},
		"coinbase":      &Coinbase{Pool: httpWorkerPool},
		"coinbasepro":   &CoinbasePro{Pool: httpWorkerPool},
		"cryptocompare": &CryptoCompare{Pool: httpWorkerPool},
		"ddex":          &Ddex{Pool: httpWorkerPool},
		"folgory":       &Folgory{Pool: httpWorkerPool},
		"ftx":           &Ftx{Pool: httpWorkerPool},
		"fx":            &Fx{Pool: httpWorkerPool},
		"gateio":        &Gateio{Pool: httpWorkerPool},
		"gemini":        &Gemini{Pool: httpWorkerPool},
		"hitbtc":        &Hitbtc{Pool: httpWorkerPool},
		"huobi":         &Huobi{Pool: httpWorkerPool},
		"kraken":        &Kraken{Pool: httpWorkerPool},
		"kucoin":        &Kucoin{Pool: httpWorkerPool},
		"kyber":         &Kyber{Pool: httpWorkerPool},
		"loopring":      &Loopring{Pool: httpWorkerPool},
		"okex":          &Okex{Pool: httpWorkerPool},
		"poloniex":      &Poloniex{Pool: httpWorkerPool},
		"uniswap":       &Uniswap{Pool: httpWorkerPool},
		"upbit":         &Upbit{Pool: httpWorkerPool},
	})
}

// Call makes exchange call
func (e *Set) Call(ppps []*model.PotentialPricePoint) ([]*model.PricePoint, error) {
	// validate potential price points:
	for _, ppp := range ppps {
		err := model.ValidatePotentialPricePoint(ppp)
		if err != nil {
			return nil, err
		}
	}

	// group potential price points by exchange:
	pppMap := map[*model.Exchange][]*model.PotentialPricePoint{}
	for _, ppp := range ppps {
		if _, ok := pppMap[ppp.Exchange]; !ok {
			pppMap[ppp.Exchange] = []*model.PotentialPricePoint{}
		}

		pppMap[ppp.Exchange] = append(pppMap[ppp.Exchange], ppp)
	}

	// fetch data from exchanges:
	var retPP []*model.PricePoint
	for exchange, exPPPs := range pppMap {
		handler, ok := e.list[exchange.Name]
		if !ok {
			return nil, fmt.Errorf("%w (%s)", errUnknownExchange, exchange.Name)
		}

		pp, err := handler.Call(exPPPs)
		if err != nil {
			return nil, err
		}

		retPP = append(retPP, pp...)
	}

	return retPP, nil
}
