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
	"strings"
	"time"

	"github.com/makerdao/gofer/internal/query"
)

// BitTrex URL
const bittrexURL = "https://api.bittrex.com/api/v1.1/public/getticker?market=%s"

type bittrexResponse struct {
	Success bool `json:"success"`
	Result  struct {
		Ask  float64 `json:"Ask"`
		Bid  float64 `json:"Bid"`
		Last float64 `json:"Last"`
	} `json:"result"`
}

// BitTrex origin handler
type BitTrex struct {
	Pool query.WorkerPool
}

func (b *BitTrex) localPairName(pair Pair) string {
	return fmt.Sprintf("%s-%s", strings.ToUpper(pair.Quote), strings.ToUpper(pair.Base))
}

func (b *BitTrex) getURL(pair Pair) string {
	return fmt.Sprintf(bittrexURL, b.localPairName(pair))
}

func (b *BitTrex) Fetch(pairs []Pair) []FetchResult {
	return callSinglePairOrigin(b, pairs)
}

func (b *BitTrex) callOne(pair Pair) (*Tick, error) {
	var err error
	req := &query.HTTPRequest{
		URL: b.getURL(pair),
	}

	// make query
	res := b.Pool.Query(req)
	if res == nil {
		return nil, errEmptyOriginResponse
	}
	if res.Error != nil {
		return nil, res.Error
	}
	// parsing JSON
	var resp bittrexResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bittrex response: %w", err)
	}
	if !resp.Success {
		return nil, fmt.Errorf("wrong response from bittrex %v", resp)
	}
	// building Tick
	return &Tick{
		Pair:      pair,
		Price:     resp.Result.Last,
		Ask:       resp.Result.Ask,
		Bid:       resp.Result.Bid,
		Timestamp: time.Now(),
	}, nil
}
