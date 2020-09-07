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

// Handler is interface that all Exchange API handlers should implement.
type Handler interface {
	// Call should implement making API request to exchange URL and
	// collecting/parsing exchange data.
	Fetch(ppps []*model.PricePoint)
}

type pppList []*model.PricePoint

// groupByHandler groups Price Points by handler. Grouped PPPs are returned
// as a map where the key is the handler and value is the list of all PPs
// assigned to that handler.
func (p pppList) groupByHandler() map[*model.Exchange]pppList {
	pppMap := map[*model.Exchange]pppList{}
	for _, ppp := range p {
		if _, ok := pppMap[ppp.Exchange]; !ok {
			pppMap[ppp.Exchange] = []*model.PricePoint{}
		}
		pppMap[ppp.Exchange] = append(pppMap[ppp.Exchange], ppp)
	}

	return pppMap
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

// Call makes handler call using handlers from the Set structure.
func (e *Set) Fetch(ppps []*model.PricePoint) {
	var err error

	// The loop below uses pppList.groupByHandler method to group all PPPs.
	// Grouping allows to pass multiple PPPs to one handler which helps
	// to reduce the number of API calls.
	for exchange, ppps := range pppList(ppps).groupByHandler() {
		err = model.ValidateExchange(exchange)
		if err != nil {
			for _, ppp := range ppps {
				ppp.Error = err
			}
		} else {
			handler, ok := e.list[exchange.Name]
			if !ok {
				err = fmt.Errorf("%w (%s)", errUnknownExchange, exchange.Name)
				for _, ppp := range ppps {
					ppp.Error = err
				}
			} else {
				handler.Fetch(ppps)
			}
		}
	}
}
