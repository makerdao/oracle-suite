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

package origins

import (
	"fmt"
	"time"

	"github.com/makerdao/gofer/internal/query"
)

// Handler is interface that all Origin API handlers should implement.
type Handler interface {
	// Fetch should implement making API request to origin URL and
	// collecting/parsing origin data.
	Fetch(pairs []Pair) []FetchResult
}

type Pair struct {
	Quote string
	Base  string
}

func (p Pair) String() string {
	return fmt.Sprintf("%s/%s", p.Base, p.Quote)
}

type Tick struct {
	Pair      Pair
	Price     float64
	Bid       float64
	Ask       float64
	Volume24h float64
	Timestamp time.Time
}

type FetchResult struct {
	Tick  Tick
	Error error
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
func (e *Set) Call(originPairs map[string][]Pair) map[string][]FetchResult {
	var err error

	crs := map[string][]FetchResult{}
	for origin, pairs := range originPairs {
		handler, ok := e.list[origin]
		if !ok {
			err = fmt.Errorf("%w (%s)", errUnknownOrigin, origin)
			for _, pair := range pairs {
				crs[origin] = append(crs[origin], FetchResult{
					Tick:  Tick{Pair: pair},
					Error: err,
				})
			}
		} else {
			for _, cr := range handler.Fetch(pairs) {
				crs[origin] = append(crs[origin], cr)
			}
		}
	}

	return crs
}

type singlePairOrigin interface {
	callOne(pair Pair) (*Tick, error)
}

func callSinglePairOrigin(e singlePairOrigin, pairs []Pair) []FetchResult {
	crs := make([]FetchResult, 0)
	for _, pair := range pairs {
		tick, err := e.callOne(pair)
		if err != nil {
			crs = append(crs, FetchResult{
				Tick:  Tick{Pair: pair},
				Error: err,
			})
		} else {
			crs = append(crs, FetchResult{
				Tick:  *tick,
				Error: err,
			})
		}
	}

	return crs
}
