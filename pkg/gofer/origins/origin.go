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
	"github.com/makerdao/oracle-suite/pkg/ethereum"
)

// Handler is interface that all Origin API handlers should implement.
type Handler interface {
	// Fetch should implement making API request to origin URL and
	// collecting/parsing origin data.
	Fetch(pairs []Pair) []FetchResult
}

type ExchangeHandler interface {
	// PullPrices is similar to Handler.Fetch
	// but pairs will be already renamed based on given BaseExchangeHandler.symbolAliases
	PullPrices(pairs []Pair) []FetchResult
}

type BaseExchangeHandler struct {
	ExchangeHandler
	aliases SymbolAliases
}

func NewBaseExchangeHandler(handler ExchangeHandler, aliases SymbolAliases) *BaseExchangeHandler {
	return &BaseExchangeHandler{
		ExchangeHandler: handler,
		aliases:         aliases,
	}
}

func (h BaseExchangeHandler) Fetch(pairs []Pair) []FetchResult {
	if h.aliases == nil {
		return h.PullPrices(pairs)
	}

	var renamedPairs []Pair
	for _, pair := range pairs {
		renamedPairs = append(renamedPairs, h.aliases.replacePair(pair))
	}
	results := h.PullPrices(renamedPairs)

	// Reverting our replacement
	for i := range results {
		results[i].Price.Pair = h.aliases.revertPair(results[i].Price.Pair)
	}
	return results
}

type ContractAddresses map[string]string

func (c ContractAddresses) ByPair(p Pair) (string, bool, bool) {
	var inverted bool
	contract, ok := c[fmt.Sprintf("%s/%s", p.Base, p.Quote)]
	if !ok {
		inverted = true
		contract, ok = c[fmt.Sprintf("%s/%s", p.Quote, p.Base)]
	}
	return contract, inverted, ok
}

func (c ContractAddresses) AddressByPair(pair Pair) (ethereum.Address, bool, error) {
	contract, inverted, ok := c.ByPair(pair)
	if !ok {
		return ethereum.Address{}, inverted, fmt.Errorf("failed to get contract address for pair: %s", pair.String())
	}
	return ethereum.HexToAddress(contract), inverted, nil
}

type SymbolAliases map[string]string

func (a SymbolAliases) replaceSymbol(symbol string) (string, bool) {
	replacement, ok := a[symbol]
	if !ok {
		return symbol, false
	}
	return replacement, ok
}

// revertSymbol reverts symbol replacement.
func (a SymbolAliases) revertSymbol(symbol string) string {
	for pre, post := range a {
		if symbol == post {
			return pre
		}
	}

	return symbol
}

func (a SymbolAliases) replacePair(pair Pair) Pair {
	base, baseOk := a.replaceSymbol(pair.Base)
	quote, quoteOk := a.replaceSymbol(pair.Quote)

	return Pair{Base: base, Quote: quote, baseReplaced: baseOk, quoteReplaced: quoteOk}
}

func (a SymbolAliases) revertPair(pair Pair) Pair {
	base := pair.Base
	if pair.baseReplaced {
		base = a.revertSymbol(pair.Base)
	}

	quote := pair.Quote
	if pair.quoteReplaced {
		quote = a.revertSymbol(pair.Quote)
	}

	return Pair{Base: base, Quote: quote, baseReplaced: false, quoteReplaced: false}
}

type Pair struct {
	Base          string
	Quote         string
	baseReplaced  bool
	quoteReplaced bool
}

func (p Pair) String() string {
	return fmt.Sprintf("%s/%s", p.Base, p.Quote)
}
func (p Pair) Inverse() Pair {
	return Pair{
		Base:          p.Quote,
		Quote:         p.Base,
		baseReplaced:  p.quoteReplaced,
		quoteReplaced: p.baseReplaced,
	}
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

type Set struct {
	list map[string]Handler
}

func NewSet(list map[string]Handler) *Set {
	return &Set{list: list}
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
		origin, pairs := origin, pairs
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

func DefaultOriginSet(pool query.WorkerPool) *Set {
	return NewSet(map[string]Handler{
		"binance":       NewBaseExchangeHandler(Binance{WorkerPool: pool}, nil),
		"bitfinex":      NewBaseExchangeHandler(Bitfinex{WorkerPool: pool}, nil),
		"bitstamp":      NewBaseExchangeHandler(Bitstamp{WorkerPool: pool}, nil),
		"bitthumb":      NewBaseExchangeHandler(BitThump{WorkerPool: pool}, nil),
		"bithumb":       NewBaseExchangeHandler(BitThump{WorkerPool: pool}, nil),
		"coinbase":      NewBaseExchangeHandler(CoinbasePro{WorkerPool: pool}, nil),
		"coinbasepro":   NewBaseExchangeHandler(CoinbasePro{WorkerPool: pool}, nil),
		"cryptocompare": NewBaseExchangeHandler(CryptoCompare{WorkerPool: pool}, nil),
		"ddex":          NewBaseExchangeHandler(Ddex{WorkerPool: pool}, nil),
		"folgory":       NewBaseExchangeHandler(Folgory{WorkerPool: pool}, nil),
		"ftx":           NewBaseExchangeHandler(Ftx{WorkerPool: pool}, nil),
		"gateio":        NewBaseExchangeHandler(Gateio{WorkerPool: pool}, nil),
		"gemini":        NewBaseExchangeHandler(Gemini{WorkerPool: pool}, nil),
		"hitbtc":        NewBaseExchangeHandler(Hitbtc{WorkerPool: pool}, nil),
		"huobi":         NewBaseExchangeHandler(Huobi{WorkerPool: pool}, nil),
		"kraken":        NewBaseExchangeHandler(Kraken{WorkerPool: pool}, nil),
		"kucoin":        NewBaseExchangeHandler(Kucoin{WorkerPool: pool}, nil),
		"kyber":         NewBaseExchangeHandler(Kyber{WorkerPool: pool}, nil),
		"loopring":      NewBaseExchangeHandler(Loopring{WorkerPool: pool}, nil),
		"okex":          NewBaseExchangeHandler(Okex{WorkerPool: pool}, nil),
		"upbit":         NewBaseExchangeHandler(Upbit{WorkerPool: pool}, nil),
	})
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

type jsonrpcMessage struct {
	Version string          `json:"jsonrpc,omitempty"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Error   *jsonError      `json:"error,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
}

func (msg *jsonrpcMessage) isCall() bool {
	return msg.hasValidID() && msg.Method != ""
}

func (msg *jsonrpcMessage) isResponse() bool {
	return msg.hasValidID() && msg.Method == "" && msg.Params == nil && (msg.Result != nil || msg.Error != nil)
}

func (msg *jsonrpcMessage) hasValidID() bool {
	return len(msg.ID) > 0 && msg.ID[0] != '{' && msg.ID[0] != '['
}

func (msg *jsonrpcMessage) String() string {
	b, _ := json.Marshal(msg)
	return string(b)
}

type jsonError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (err *jsonError) Error() string {
	if err.Message == "" {
		return fmt.Sprintf("json-rpc error %d", err.Code)
	}
	return err.Message
}

func (err *jsonError) ErrorCode() int {
	return err.Code
}
