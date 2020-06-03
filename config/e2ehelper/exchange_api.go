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

package e2ehelper

import (
	"github.com/makerdao/gofer/exchange"
	"github.com/makerdao/gofer/model"
	"github.com/makerdao/gofer/query"
)

type FakeWorkerPool struct{}

func (mwp *FakeWorkerPool) Ready() bool {
	return true
}

func (mwp *FakeWorkerPool) Start() {}

func (mwp *FakeWorkerPool) Stop() error {
	return nil
}

func (mwp *FakeWorkerPool) Query(req *query.HTTPRequest) *query.HTTPResponse {
	ethUsd := &model.PotentialPricePoint{
		Pair: &model.Pair{
			Base:  "ETH",
			Quote: "USD",
		},
	}

	batUsd := &model.PotentialPricePoint{
		Pair: &model.Pair{
			Base:  "ETH",
			Quote: "USD",
		},
	}

	bitstamp := &exchange.Bitstamp{}
	coinbase := &exchange.Coinbase{}
	gemini := &exchange.Gemini{}
	kraken := &exchange.Kraken{}
	binance := &exchange.Binance{}
	bittrex := &exchange.BitTrex{}
	upbit := &exchange.Upbit{}

	body := "{\"error\":\"unknown\"}"
	switch req.URL {
	case bitstamp.GetURL(ethUsd):
		body = "{\"high\":\"252.29\",\"last\":\"239.51\",\"timestamp\":\"1591191281\",\"bid\":\"239.30\",\"vwap\":\"236.70\",\"volume\":\"111520.55395880\",\"low\":\"225.00\",\"ask\":\"239.52\",\"open\":\"237.81\"}"
	case coinbase.GetURL(ethUsd):
		body = "{\"trade_id\":58917061,\"price\":\"239.71\",\"size\":\"0.19098533\",\"time\":\"2020-06-03T13:48:34.8134Z\",\"bid\":\"239.67\",\"ask\":\"239.68\",\"volume\":\"232589.71348494\"}"
	case gemini.GetURL(ethUsd):
		body = "{\"bid\":\"239.64\",\"ask\":\"239.82\",\"volume\":{\"ETH\":\"27270.82577847\",\"USD\":\"6433027.9072109886\",\"timestamp\":1591191900000},\"last\":\"239.50\"}"
	case kraken.GetURL(ethUsd):
		body = "{\"error\":[],\"result\":{\"XETHZUSD\":{\"a\":[\"239.76000\",\"4\",\"4.000\"],\"b\":[\"239.75000\",\"1\",\"1.000\"],\"c\":[\"239.74000\",\"0.19083179\"],\"v\":[\"31378.67589429\",\"106844.51178263\"],\"p\":[\"237.40044\",\"236.61147\"],\"t\":[4495,15800],\"l\":[\"233.47000\",\"218.48000\"],\"h\":[\"240.70000\",\"252.29000\"],\"o\":\"237.55000\"}}}"
	case binance.GetURL(batUsd):
		body = ""
	case bittrex.GetURL(batUsd):
		body = ""
	case coinbase.GetURL(batUsd):
		body = ""
	case upbit.GetURL(batUsd):
		body = ""
	}

	return &query.HTTPResponse{
		Body: []byte(body),
	}
}

// NewETHUSDMockWorkerPool is made for testing purposes only.
// it will create new fake worker pool that will handle only several requests
// to concrete exchanges:
// ETHUSD:
// - bitstamp
// - coinbase
// - gemini
// - kraken
//
// BATUSD
// - binance
// - bittrex
// - coinbase
// - upbit
func NewFakeWorkerPool() query.WorkerPool {
	return &FakeWorkerPool{}
}
