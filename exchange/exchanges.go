package exchange

import (
	"makerdao/gofer/model"
	"makerdao/gofer/query"
)

// Handler is interface that all Exchange API handlers should implement
type Handler interface {
	// Call should implement making API request to exchange URL and collecting/parsing exhcange data
	Call(pool query.WorkerPool, pp *model.PotentialPricePoint) (*model.PricePoint, error)
}

// List of implemented exchanges
var exchangeList = map[string]Handler{
	"binance":     &Binance{},
	"bitfinex":    &Bitfinex{},
	"bitstamp":    &Bitstamp{},
	"bittrex":     &BitTrex{},
	"coinbase":    &Coinbase{},
	"coinbasepro": &CoinbasePro{},
	"fx":          &Fx{},
	"gateio":      &Gateio{},
	"gemini":      &Gemini{},
	"hitbtc":      &Hitbtc{},
	"huobi":       &Huobi{},
	"poloniex":    &Poloniex{},
	"upbit":       &Upbit{},
}

// Call makes exchange call
func Call(pool query.WorkerPool, pp *model.PotentialPricePoint) (*model.PricePoint, error) {
	if pool == nil {
		return nil, errNoPoolPassed
	}
	if pp == nil {
		return nil, errNoPotentialPricePoint
	}
	err := model.ValidatePotentialPricePoint(pp)
	if err != nil {
		return nil, err
	}

	handler, ok := exchangeList[pp.Exchange.Name]
	if !ok {
		return nil, errUnknownExchange
	}
	return handler.Call(pool, pp)
}
