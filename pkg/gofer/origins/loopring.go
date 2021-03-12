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

// Loopring URL
const loopringURL = "https://api.loopring.io/api/v2/overview"

type pairResponse struct {
	Price  stringAsFloat64 `json:"last"`
	Ask    stringAsFloat64 `json:"lowestAsk"`
	Bid    stringAsFloat64 `json:"highestBid"`
	Volume stringAsFloat64 `json:"quoteVolume"`
}

type loopringResponse struct {
	ResultInfo struct {
		Code    int
		Message string
	} `json:"resultInfo"`
	Data map[string]pairResponse `json:"data"`
}

// Loopring origin handler
type Loopring struct {
	Pool query.WorkerPool
}

func (l *Loopring) localPairName(pair Pair) string {
	return fmt.Sprintf("%s-%s", strings.ToUpper(pair.Base), strings.ToUpper(pair.Quote))
}

func (l *Loopring) getURL() string {
	return loopringURL
}

func (l *Loopring) Fetch(pairs []Pair) []FetchResult {
	var err error
	req := &query.HTTPRequest{
		URL: l.getURL(),
	}
	// make query
	res := l.Pool.Query(req)
	if res == nil {
		return fetchResultListWithErrors(pairs, ErrEmptyOriginResponse)
	}
	if res.Error != nil {
		return fetchResultListWithErrors(pairs, res.Error)
	}
	// parsing JSON
	var resp loopringResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		return fetchResultListWithErrors(pairs, fmt.Errorf("failed to parse loopring response: %w", err))
	}
	if resp.ResultInfo.Code != 0 || resp.ResultInfo.Message != "SUCCESS" {
		return fetchResultListWithErrors(pairs, fmt.Errorf("wrong loopring response %s", res.Body))
	}
	if resp.Data == nil {
		return fetchResultListWithErrors(pairs, fmt.Errorf("empty `data` field for loopring response: %s", res.Body))
	}

	var results []FetchResult
	for _, pair := range pairs {
		results = append(results, l.pickPairDetails(resp, pair))
	}
	return results
}

func (l *Loopring) pickPairDetails(response loopringResponse, pair Pair) FetchResult {
	pairName := l.localPairName(pair)
	pairRes, ok := response.Data[pairName]
	if !ok {
		return fetchResultWithError(pair, fmt.Errorf("no %s pair exist in loopring response", pair))
	}
	// building Price
	return fetchResult(Price{
		Pair:      pair,
		Price:     pairRes.Price.val(),
		Volume24h: pairRes.Volume.val(),
		Ask:       pairRes.Ask.val(),
		Bid:       pairRes.Bid.val(),
		Timestamp: time.Now(),
	})
}
