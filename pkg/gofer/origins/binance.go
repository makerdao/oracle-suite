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

	"github.com/makerdao/gofer/internal/query"
)

const binanceURL = "https://www.binance.com/api/v3/ticker/24hr"

type binanceResponse struct {
	Symbol    string               `json:"symbol"`
	LastPrice stringAsFloat64      `json:"lastPrice"`
	BidPrice  stringAsFloat64      `json:"bidPrice"`
	AskPrice  stringAsFloat64      `json:"askPrice"`
	Volume    stringAsFloat64      `json:"volume"`
	CloseTime intAsUnixTimestampMs `json:"closeTime"`
}

// Binance origin handler
type Binance struct {
	Pool query.WorkerPool
}

func (b *Binance) localPairName(pair Pair) string {
	return pair.Base + pair.Quote
}

func (b *Binance) Fetch(pairs []Pair) []FetchResult {
	var err error
	req := &query.HTTPRequest{
		URL: binanceURL,
	}

	// make query
	res := b.Pool.Query(req)
	if res == nil {
		return fetchResultListWithErrors(pairs, errEmptyOriginResponse)
	}
	if res.Error != nil {
		return fetchResultListWithErrors(pairs, res.Error)
	}

	// parse JSON
	var resp []binanceResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		return fetchResultListWithErrors(pairs, fmt.Errorf("failed to parse Binance response: %w", err))
	}

	// convert response from a slice to a map
	respMap := map[string]binanceResponse{}
	for _, symbolResp := range resp {
		respMap[symbolResp.Symbol] = symbolResp
	}

	// prepare result
	results := make([]FetchResult, 0)
	for _, pair := range pairs {
		if r, ok := respMap[b.localPairName(pair)]; !ok {
			results = append(results, FetchResult{
				Tick:  Tick{Pair: pair},
				Error: errMissingResponseForPair,
			})
		} else {
			results = append(results, FetchResult{
				Tick: Tick{
					Pair:      pair,
					Price:     r.LastPrice.val(),
					Bid:       r.BidPrice.val(),
					Ask:       r.AskPrice.val(),
					Volume24h: r.Volume.val(),
					Timestamp: r.CloseTime.val(),
				},
			})
		}
	}

	return results
}
