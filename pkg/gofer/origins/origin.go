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
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/makerdao/oracle-suite/internal/query"
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

func (p Pair) Equal(c Pair) bool {
	return p.Base == c.Base && p.Quote == c.Quote
}

type Price struct {
	Pair      Pair
	Price     float64
	Bid       float64
	Ask       float64
	Volume24h float64
	Timestamp time.Time
}

type FetchResult struct {
	Price Price
	Error error
}

func fetchResult(price Price) FetchResult {
	return FetchResult{
		Price: price,
		Error: nil,
	}
}

func fetchResultWithError(pair Pair, err error) FetchResult {
	return FetchResult{
		Price: Price{
			Pair:      pair,
			Timestamp: time.Now(),
		},
		Error: err,
	}
}

func fetchResultListWithErrors(pairs []Pair, err error) []FetchResult {
	r := make([]FetchResult, len(pairs))
	for i, pair := range pairs {
		r[i] = FetchResult{
			Price: Price{
				Pair:      pair,
				Timestamp: time.Now(),
			},
			Error: err,
		}
	}
	return r
}

//nolint
func NewHandler(handlerType string, pool query.WorkerPool, params json.RawMessage) Handler {
	switch handlerType {
	case "balancer":
		return &Balancer{Pool: pool}
	case "binance":
		return &Binance{Pool: pool}
	case "bitfinex":
		return &Bitfinex{Pool: pool}
	case "bitstamp":
		return &Bitstamp{Pool: pool}
	case "bitthumb":
		return &BitThump{Pool: pool}
	case "bithumb":
		return &BitThump{Pool: pool}
	case "bittrex":
		return &Bittrex{Pool: pool}
	case "coinbase", "coinbasepro":
		return &CoinbasePro{Pool: pool}
	case "cryptocompare":
		return &CryptoCompare{Pool: pool}
	case "coinmarketcap":
		return NewCoinMarketCap(pool, params)
	case "ddex":
		return &Ddex{Pool: pool}
	case "folgory":
		return &Folgory{Pool: pool}
	case "ftx":
		return &Ftx{Pool: pool}
	case "fx":
		return &Fx{Pool: pool}
	case "gateio":
		return &Gateio{Pool: pool}
	case "gemini":
		return &Gemini{Pool: pool}
	case "hitbtc":
		return &Hitbtc{Pool: pool}
	case "huobi":
		return &Huobi{Pool: pool}
	case "kraken":
		return &Kraken{Pool: pool}
	case "kucoin":
		return &Kucoin{Pool: pool}
	case "kyber":
		return &Kyber{Pool: pool}
	case "loopring":
		return &Loopring{Pool: pool}
	case "okex":
		return &Okex{Pool: pool}
	case "operexchangerates":
		return NewOpenExchangeRates(pool, params)
	case "poloniex":
		return &Poloniex{Pool: pool}
	case "sushiswap":
		return &Sushiswap{Pool: pool}
	case "uniswap":
		return &Uniswap{Pool: pool}
	case "upbit":
		return &Upbit{Pool: pool}
	}

	return nil
}

type Set struct {
	list map[string]Handler
}

func NewSet(list map[string]Handler) *Set {
	return &Set{list: list}
}

func DefaultSet(httpWorkerPool query.WorkerPool) *Set {
	return NewSet(map[string]Handler{
		"balancer":          NewHandler("balancer", httpWorkerPool, nil),
		"binance":           NewHandler("binance", httpWorkerPool, nil),
		"bitfinex":          NewHandler("bitfinex", httpWorkerPool, nil),
		"bitstamp":          NewHandler("bitstamp", httpWorkerPool, nil),
		"bitthumb":          NewHandler("bitthumb", httpWorkerPool, nil),
		"bithumb":           NewHandler("bithumb", httpWorkerPool, nil),
		"bittrex":           NewHandler("bittrex", httpWorkerPool, nil),
		"coinbasepro":       NewHandler("coinbasepro", httpWorkerPool, nil),
		"cryptocompare":     NewHandler("cryptocompare", httpWorkerPool, nil),
		"coinmarketcap":     NewHandler("coinmarketcap", httpWorkerPool, nil),
		"ddex":              NewHandler("ddex", httpWorkerPool, nil),
		"folgory":           NewHandler("folgory", httpWorkerPool, nil),
		"ftx":               NewHandler("ftx", httpWorkerPool, nil),
		"fx":                NewHandler("fx", httpWorkerPool, nil),
		"gateio":            NewHandler("gateio", httpWorkerPool, nil),
		"gemini":            NewHandler("gemini", httpWorkerPool, nil),
		"hitbtc":            NewHandler("hitbtc", httpWorkerPool, nil),
		"huobi":             NewHandler("huobi", httpWorkerPool, nil),
		"kraken":            NewHandler("kraken", httpWorkerPool, nil),
		"kucoin":            NewHandler("kucoin", httpWorkerPool, nil),
		"kyber":             NewHandler("kyber", httpWorkerPool, nil),
		"loopring":          NewHandler("loopring", httpWorkerPool, nil),
		"okex":              NewHandler("okex", httpWorkerPool, nil),
		"operexchangerates": NewHandler("operexchangerates", httpWorkerPool, nil),
		"poloniex":          NewHandler("poloniex", httpWorkerPool, nil),
		"sushiswap":         NewHandler("sushiswap", httpWorkerPool, nil),
		"uniswap":           NewHandler("uniswap", httpWorkerPool, nil),
		"upbit":             NewHandler("upbit", httpWorkerPool, nil),
	})
}

func (e *Set) GetHandler(name string) Handler {
	handler, ok := e.list[name]
	if !ok {
		return nil
	}
	return handler
}

func (e *Set) SetHandler(name string, handler Handler) {
	e.list[name] = handler
}

func (e *Set) Handlers() map[string]Handler {
	c := map[string]Handler{}
	for k, v := range e.list {
		c[k] = v
	}
	return c
}

// Fetch makes handler fetch using handlers from the Set structure.
func (e *Set) Fetch(originPairs map[string][]Pair) map[string][]FetchResult {
	var mu sync.Mutex
	var wg sync.WaitGroup

	wg.Add(len(originPairs))

	frs := map[string][]FetchResult{}
	for origin, pairs := range originPairs {
		origin := origin
		pairs := pairs
		handler, ok := e.list[origin]

		go func() {
			if !ok {
				mu.Lock()
				frs[origin] = fetchResultListWithErrors(
					pairs,
					fmt.Errorf("%w (%s)", ErrUnknownOrigin, origin),
				)
				mu.Unlock()
			} else {
				resp := handler.Fetch(pairs)
				mu.Lock()
				frs[origin] = append(frs[origin], resp...)
				mu.Unlock()
			}

			wg.Done()
		}()
	}

	wg.Wait()
	return frs
}

type singlePairOrigin interface {
	callOne(pair Pair) (*Price, error)
}

func callSinglePairOrigin(e singlePairOrigin, pairs []Pair) []FetchResult {
	crs := make([]FetchResult, 0)
	for _, pair := range pairs {
		price, err := e.callOne(pair)
		if err != nil {
			crs = append(crs, FetchResult{
				Price: Price{Pair: pair},
				Error: err,
			})
		} else {
			crs = append(crs, FetchResult{
				Price: *price,
				Error: err,
			})
		}
	}

	return crs
}

func validateResponse(pairs []Pair, res *query.HTTPResponse) []FetchResult {
	if res == nil {
		return fetchResultListWithErrors(pairs, ErrInvalidResponseStatus)
	}
	if res.Error != nil {
		return fetchResultListWithErrors(pairs, fmt.Errorf("bad response: %w", res.Error))
	}
	return nil
}
