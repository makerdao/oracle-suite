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
	"folgory":     &Folgory{},
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
