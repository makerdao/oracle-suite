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
	"time"

	"github.com/makerdao/gofer/internal/query"
)

const okexURL = "https://www.okex.com/api/spot/v3/instruments/ticker"

type okexResponse struct {
	InstrumentID  string          `json:"instrument_id"`
	Last          stringAsFloat64 `json:"last"`
	BestAsk       stringAsFloat64 `json:"best_ask"`
	BestBid       stringAsFloat64 `json:"best_bid"`
	BaseVolume24H stringAsFloat64 `json:"base_volume_24h"`
	Timestamp     time.Time       `json:"timestamp"`
}

// Okex origin handler
type Okex struct {
	Pool query.WorkerPool
}

func (o *Okex) localPairName(pair Pair) string {
	return fmt.Sprintf("%s-%s", pair.Base, pair.Quote)
}

func (o *Okex) Fetch(pairs []Pair) []FetchResult {
	var err error
	req := &query.HTTPRequest{
		URL: okexURL,
	}

	// make query
	res := o.Pool.Query(req)
	if res == nil {
		return fetchResultListWithErrors(pairs, errEmptyOriginResponse)
	}
	if res.Error != nil {
		return fetchResultListWithErrors(pairs, res.Error)
	}

	// parse JSON
	var resp []okexResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		return fetchResultListWithErrors(pairs, fmt.Errorf("failed to parse Okex response: %w", err))
	}

	// convert response from a slice to a map
	respMap := map[string]okexResponse{}
	for _, symbolResp := range resp {
		respMap[symbolResp.InstrumentID] = symbolResp
	}

	// prepare result
	results := make([]FetchResult, 0)
	for _, pair := range pairs {
		if r, ok := respMap[o.localPairName(pair)]; !ok {
			results = append(results, FetchResult{
				Tick:  Tick{Pair: pair},
				Error: errMissingResponseForPair,
			})
		} else {
			results = append(results, FetchResult{
				Tick: Tick{
					Pair:      pair,
					Price:     r.Last.val(),
					Bid:       r.BestBid.val(),
					Ask:       r.BestAsk.val(),
					Volume24h: r.BaseVolume24H.val(),
					Timestamp: r.Timestamp,
				},
			})
		}
	}

	return results
}
