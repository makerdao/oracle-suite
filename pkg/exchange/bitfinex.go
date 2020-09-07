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
	"encoding/json"
	"fmt"
	"time"

	"github.com/makerdao/gofer/internal/query"
	"github.com/makerdao/gofer/pkg/model"
)

// Bitfinex URL
const bitfinexURL = "https://api-pub.bitfinex.com/v2/ticker/t%s"

// Bitfinex exchange handler
type Bitfinex struct {
	Pool query.WorkerPool
}

func (b *Bitfinex) localPairName(pair *model.Pair) string {
	const USDT = "USDT"
	const USD = "USD"
	if pair.Base == USDT && pair.Quote == USD {
		return "USTUSD"
	}
	if pair.Quote == USDT {
		return pair.Base + USD
	}
	return pair.Base + pair.Quote
}

func (b *Bitfinex) getURL(pp *model.PricePoint) string {
	var pair string

	if pp.Exchange == nil {
		pair = b.localPairName(pp.Pair)
	} else {
		pair = pp.Exchange.Config["pair"]
	}
	if pair == "" {
		pair = b.localPairName(pp.Pair)
	}
	return fmt.Sprintf(bitfinexURL, pair)
}

func (b *Bitfinex) Fetch(pps []*model.PricePoint) {
	for _, pp := range pps {
		b.fetchOne(pp)
	}
}

func (b *Bitfinex) fetchOne(pp *model.PricePoint) {
	err := model.ValidatePricePoint(pp)
	if err != nil {
		pp.Error = err
		return
	}

	req := &query.HTTPRequest{
		URL: b.getURL(pp),
	}

	// make query
	res := b.Pool.Query(req)
	if res == nil {
		pp.Error = errEmptyExchangeResponse
		return
	}
	if res.Error != nil {
		pp.Error = res.Error
		return
	}

	// parsing JSON
	var resp []float64
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		pp.Error = fmt.Errorf("failed to parse bitfinex response: %w", err)
		return
	}
	if len(resp) < 8 {
		pp.Error = fmt.Errorf("wrong bitfinex response")
		return
	}

	pp.Timestamp = time.Now().Unix()
	pp.Price = resp[6]
	pp.Volume = resp[7]
}
