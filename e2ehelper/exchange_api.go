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

func newPotentialPoint(base, quote string) *model.PotentialPricePoint {
	return &model.PotentialPricePoint{
		Pair: model.NewPair(base, quote),
	}
}

//TODO: add BTC/USD `minimum amount of sources not met for 'BTC/USD' in price model`
func (mwp *FakeWorkerPool) Query(req *query.HTTPRequest) *query.HTTPResponse {
	ethUsd := newPotentialPoint("ETH", "USD")
	ethBtc := newPotentialPoint("ETH", "BTC")
	ethUsdt := newPotentialPoint("ETH", "USDT")
	batUsd := newPotentialPoint("BAT", "USD")
	batBtc := newPotentialPoint("BAT", "BTC")
	btcUsd := newPotentialPoint("BTC", "USD")
	batKrw := newPotentialPoint("BAT", "KRW")
	krwUsd := newPotentialPoint("KRW", "USD")

	bitstamp := &exchange.Bitstamp{}
	coinbase := &exchange.Coinbase{}
	gemini := &exchange.Gemini{}
	kraken := &exchange.Kraken{}
	binance := &exchange.Binance{}
	bittrex := &exchange.BitTrex{}
	bitfinex := &exchange.Bitfinex{}
	upbit := &exchange.Upbit{}
	fx := &exchange.Fx{}

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
	case binance.GetURL(batBtc):
		body = "{\"symbol\":\"BATBTC\",\"price\":\"0.00002489\"}"
	case binance.GetURL(ethBtc):
		body = "{\"symbol\":\"ETHBTC\",\"price\":\"0.03298700\"}"
	case bitfinex.GetURL(ethUsdt):
		body = "[395.35,527.27243938,395.36,1001.1600643000002,1.10541502,0.0028,395.36,79015.64811632,400.4,382.45]"
	case bittrex.GetURL(batBtc):
		body = "{\"success\":true,\"message\":\"\",\"result\":{\"Bid\":0.00002484,\"Ask\":0.00002493,\"Last\":0.00002488}}"
	case coinbase.GetURL(batUsd):
		body = "{\"trade_id\":3575777,\"price\":\"0.243869\",\"size\":\"2\",\"time\":\"2020-06-05T08:50:05.469136Z\",\"bid\":\"0.243943\",\"ask\":\"0.244333\",\"volume\":\"12368752.00000000\"}"
	case upbit.GetURL(batKrw):
		body = "[{\"market\":\"KRW-BAT\",\"trade_date\":\"20200605\",\"trade_time\":\"085628\",\"trade_date_kst\":\"20200605\",\"trade_time_kst\":\"175628\",\"trade_timestamp\":1591347388000,\"opening_price\":291.00000000,\"high_price\":298.00000000,\"low_price\":288.00000000,\"trade_price\":292.00000000,\"prev_closing_price\":290.00000000,\"change\":\"RISE\",\"change_price\":2.00000000,\"change_rate\":0.0068965517,\"signed_change_price\":2.00000000,\"signed_change_rate\":0.0068965517,\"trade_volume\":21.40466320,\"acc_trade_price\":314088265.68810931,\"acc_trade_price_24h\":688375303.17847536,\"acc_trade_volume\":1072719.24124764,\"acc_trade_volume_24h\":2394247.57223916,\"highest_52_week_price\":440.00000000,\"highest_52_week_date\":\"2019-06-09\",\"lowest_52_week_price\":111.00000000,\"lowest_52_week_date\":\"2020-03-13\",\"timestamp\":1591347388594}]"
	case fx.GetURL(krwUsd):
		body = "{\"rates\":{\"CAD\":0.0011107377,\"HKD\":0.0063700657,\"ISK\":0.108494736,\"PHP\":0.0409934757,\"DKK\":0.0054471664,\"HUF\":0.2519854171,\"CZK\":0.0194508778,\"GBP\":0.0006552425,\"RON\":0.0035341521,\"SEK\":0.0076108509,\"IDR\":11.5851044399,\"INR\":0.0620516829,\"BRL\":0.0041636407,\"RUB\":0.0568506572,\"HRK\":0.0055325009,\"JPY\":0.0894844126,\"THB\":0.0259649456,\"CHF\":0.0007880298,\"EUR\":0.0007306043,\"MYR\":0.0035125262,\"BGN\":0.0014289159,\"TRY\":0.0055429486,\"CNY\":0.0058496563,\"NOK\":0.0077479123,\"NZD\":0.0012792881,\"ZAR\":0.0138857919,\"USD\":0.0008219298,\"MXN\":0.0178823435,\"SGD\":0.0011512862,\"AUD\":0.0011891315,\"ILS\":0.0028543248,\"KRW\":1.0,\"PLN\":0.0032418373},\"base\":\"KRW\",\"date\":\"2020-06-04\"}"
	case bittrex.GetURL(btcUsd):
		body = "{\"success\":true,\"message\":\"\",\"result\":{\"Bid\":9834.25300000,\"Ask\":9839.50000000,\"Last\":9838.66000000}}"
	case bitstamp.GetURL(btcUsd):
		body = "{\"high\": \"12080.00\", \"last\": \"11988.96\", \"timestamp\": \"1597053712\", \"bid\": \"11992.55\", \"vwap\": \"11848.56\", \"volume\": \"4564.67215002\", \"low\": \"11527.92\", \"ask\": \"12000.54\", \"open\": \"11684.60\"}"
	case coinbase.GetURL(btcUsd):
		body = "{\"trade_id\":99536988,\"price\":\"12002.91\",\"size\":\"0.00141377\",\"time\":\"2020-08-10T10:03:07.544865Z\",\"bid\":\"12002.9\",\"ask\":\"12002.91\",\"volume\":\"13339.88736845\"}"
	case gemini.GetURL(btcUsd):
		body = "{\"bid\":\"12006.56\",\"ask\":\"12006.57\",\"volume\":{\"BTC\":\"1357.5435869273\",\"USD\":\"16107568.021683016761\",\"timestamp\":1597053600000},\"last\":\"12006.57\"}"
	case kraken.GetURL(btcUsd):
		body = "{\"error\":[],\"result\":{\"XXBTZUSD\":{\"a\":[\"11999.80000\",\"1\",\"1.000\"],\"b\":[\"11999.70000\",\"4\",\"4.000\"],\"c\":[\"11999.70000\",\"0.06016520\"],\"v\":[\"3122.76429853\",\"4195.89003127\"],\"p\":[\"11973.80060\",\"11884.45755\"],\"t\":[12215,18958],\"l\":[\"11687.50000\",\"11532.50000\"],\"h\":[\"12082.90000\",\"12082.90000\"],\"o\":\"11687.50000\"}}}"
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
