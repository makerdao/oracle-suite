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

type Ddex struct {
	Pool query.WorkerPool
}

const ddexPriceersURL = "https://api.ddex.io/v4/markets/priceers"

func (o *Ddex) Fetch(pairs []Pair) []FetchResult {
	req := &query.HTTPRequest{
		URL: ddexPriceersURL,
	}
	res := o.Pool.Query(req)
	if errorResponses := validateResponse(pairs, res); len(errorResponses) > 0 {
		return errorResponses
	}
	return o.parseResponse(pairs, res)
}

func (o *Ddex) localPairName(pair Pair) string {
	return fmt.Sprintf("%s-%s", pair.Base, pair.Quote)
}

type ddexPriceer struct {
	Ask      stringAsFloat64      `json:"ask"`
	Bid      stringAsFloat64      `json:"bid"`
	High     stringAsFloat64      `json:"high"`
	Low      stringAsFloat64      `json:"low"`
	MarketID string               `json:"marketId"`
	Price    stringAsFloat64      `json:"price"`
	UpdateAt intAsUnixTimestampMs `json:"updateAt"`
	Volume   stringAsFloat64      `json:"volume"`
}
type ddexPriceersResponse struct {
	Desc   string `json:"desc"`
	Status int    `json:"status"`
	Data   struct {
		Priceers []ddexPriceer `json:"priceers"`
	} `json:"data"`
}

func (o *Ddex) parseResponse(pairs []Pair, res *query.HTTPResponse) []FetchResult {
	results := make([]FetchResult, 0)
	var resp ddexPriceersResponse
	err := json.Unmarshal(res.Body, &resp)
	if err != nil {
		return fetchResultListWithErrors(pairs, fmt.Errorf("failed to parse response: %w", err))
	}
	if resp.Status != 0 {
		return fetchResultListWithErrors(pairs, ErrInvalidResponseStatus)
	}

	priceers := make(map[string]ddexPriceer)
	for _, t := range resp.Data.Priceers {
		priceers[t.MarketID] = t
	}

	for _, pair := range pairs {
		if t, is := priceers[o.localPairName(pair)]; !is {
			results = append(results, FetchResult{
				Price: Price{Pair: pair},
				Error: ErrMissingResponseForPair,
			})
		} else {
			results = append(results, FetchResult{
				Price: Price{
					Pair:      pair,
					Price:     t.Price.val(),
					Bid:       t.Bid.val(),
					Ask:       t.Ask.val(),
					Volume24h: t.Volume.val(),
					Timestamp: t.UpdateAt.val(),
				},
			})
		}
	}
	return results
}
