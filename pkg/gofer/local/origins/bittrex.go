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

const bittrexURL = "https://api.bittrex.com/api/v1.1/public/getmarketsummaries"

type bittrexResponse struct {
	Success bool                    `json:"success"`
	Result  []bittrexSymbolResponse `json:"result"`
}

type bittrexSymbolResponse struct {
	MarketName string  `json:"MarketName"`
	Ask        float64 `json:"Ask"`
	Bid        float64 `json:"Bid"`
	Last       float64 `json:"Last"`
	Volume     float64 `json:"Volume"`
	TimeStamp  string  `json:"TimeStamp"`
}

// Bittrex origin handler
type Bittrex struct {
	Pool query.WorkerPool
}

func (b *Bittrex) localPairName(pair Pair) string {
	const (
		REP   = "REP"
		REPV2 = "REPV2"
	)

	if pair.Quote == REP {
		pair.Quote = REPV2
	}

	if pair.Base == REP {
		pair.Base = REPV2
	}

	return fmt.Sprintf("%s-%s", pair.Quote, pair.Base)
}

func (b *Bittrex) Fetch(pairs []Pair) []FetchResult {
	var err error
	req := &query.HTTPRequest{
		URL: bittrexURL,
	}

	// make query
	res := b.Pool.Query(req)
	if res == nil {
		return fetchResultListWithErrors(pairs, ErrEmptyOriginResponse)
	}
	if res.Error != nil {
		return fetchResultListWithErrors(pairs, res.Error)
	}

	// parse JSON
	var resp bittrexResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		return fetchResultListWithErrors(pairs, fmt.Errorf("failed to parse Bittrex response: %w", err))
	}
	if !resp.Success {
		return fetchResultListWithErrors(pairs, fmt.Errorf("wrong response from Bittrex %v", resp))
	}

	// convert response from a slice to a map
	respMap := map[string]bittrexSymbolResponse{}
	for _, symbolResp := range resp.Result {
		respMap[symbolResp.MarketName] = symbolResp
	}

	// prepare result
	results := make([]FetchResult, 0)
	for _, pair := range pairs {
		if r, ok := respMap[b.localPairName(pair)]; !ok {
			results = append(results, FetchResult{
				Tick:  Tick{Pair: pair},
				Error: ErrMissingResponseForPair,
			})
		} else {
			// parse timestamp
			ts, err := time.Parse("2006-01-02T15:04:05", r.TimeStamp)
			if err != nil {
				ts = time.Unix(0, 0)
			}

			results = append(results, FetchResult{
				Tick: Tick{
					Pair:      pair,
					Price:     r.Last,
					Bid:       r.Bid,
					Ask:       r.Ask,
					Volume24h: r.Volume,
					Timestamp: ts,
				},
			})
		}
	}

	return results
}
