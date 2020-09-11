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
	"strconv"
	"strings"
	"time"

	"github.com/makerdao/gofer/internal/query"
)

// Loopring URL
const loopringURL = "https://api.loopring.io/api/v2/overview"

type pairResponse struct {
	Price  string `json:"last"`
	Ask    string `json:"lowestAsk"`
	Bid    string `json:"highestBid"`
	Volume string `json:"quoteVolume"`
}

type loopringResponse struct {
	ResultInfo struct {
		Code    int
		Message string
	} `json:"resultInfo"`
	Data map[string]pairResponse `json:"data"`
}

// Loopring exchange handler
type Loopring struct {
	Pool query.WorkerPool
}

func (l *Loopring) localPairName(pair Pair) string {
	return fmt.Sprintf("%s-%s", strings.ToUpper(pair.Base), strings.ToUpper(pair.Quote))
}

func (l *Loopring) getURL(pair Pair) string {
	return loopringURL
}

func (l *Loopring) Fetch(pairs []Pair) []FetchResult {
	return callSinglePairExchange(l, pairs)
}

func (l *Loopring) callOne(pair Pair) (*Tick, error) {
	var err error
	req := &query.HTTPRequest{
		URL: l.getURL(pair),
	}
	// make query
	res := l.Pool.Query(req)
	if res == nil {
		return nil, errEmptyExchangeResponse
	}
	if res.Error != nil {
		return nil, res.Error
	}
	// parsing JSON
	var resp loopringResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse loopring response: %w", err)
	}
	if resp.ResultInfo.Code != 0 || resp.ResultInfo.Message != "SUCCESS" {
		return nil, fmt.Errorf("wrong loopring response %s", res.Body)
	}
	if resp.Data == nil {
		return nil, fmt.Errorf("empty `data` field for loopring response: %s", res.Body)
	}
	pairName := l.localPairName(pair)
	pairRes, ok := resp.Data[pairName]
	if !ok {
		return nil, fmt.Errorf("no %s pair exist in loopring response: %s", pair, res.Body)
	}
	// Parsing price from string
	price, err := strconv.ParseFloat(pairRes.Price, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse price from loopring exchange %s", res.Body)
	}
	// Parsing price from string
	volume, err := strconv.ParseFloat(pairRes.Volume, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse volume from loopring exchange %s", res.Body)
	}
	ask, err := strconv.ParseFloat(pairRes.Ask, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ask from loopring exchange %s", res.Body)
	}
	bid, err := strconv.ParseFloat(pairRes.Bid, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bid from loopring exchange %s", res.Body)
	}
	// building Tick
	return &Tick{
		Pair:      pair,
		Price:     price,
		Volume24h: volume,
		Ask:       ask,
		Bid:       bid,
		Timestamp: time.Now(),
	}, nil
}
