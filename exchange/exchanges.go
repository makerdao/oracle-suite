package exchange

import (
	"makerdao/gofer"
	"makerdao/gofer/query"
)

// Handler is interface that all Exchange API handlers should implement
type Handler interface {
	// Call should implement making API request to exchange URL and collecting/parsing exhcange data
	Call(pool *query.WorkerPool, pp *gofer.PotentialPricePoint) (*gofer.PricePoint, error)
}

// List of implemented exchanges
var exchangeList = map[string]Handler{
	"binance": &Binance{},
}

// Call makes exchange call
func Call(pool *query.WorkerPool, pp *gofer.PotentialPricePoint) (*gofer.PricePoint, error) {
	if pp == nil {
		return nil, errNoPotentialPricePoint
	}
	if pp.Exchange == nil {
		return nil, errNoExchangeInPotentialPricePoint
	}

	handler, ok := exchangeList[pp.Exchange.Name]
	if !ok {
		return nil, errUnknownExchange
	}
	return handler.Call(pool, pp)
}
