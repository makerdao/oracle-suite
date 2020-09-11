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
	"time"

	"github.com/makerdao/gofer/internal/query"
)

// Okex URL
const okexURL = "https://www.okex.com/api/spot/v3/instruments/%s/ticker"

type okexResponse struct {
	Last          string    `json:"last"`
	Ask           string    `json:"ask"`
	Bid           string    `json:"bid"`
	BaseVolume24H string    `json:"base_volume_24h"`
	Timestamp     time.Time `json:"timestamp"`
}

// Okex origin handler
type Okex struct {
	Pool query.WorkerPool
}

func (o *Okex) localPairName(pair Pair) string {
	return fmt.Sprintf("%s-%s", pair.Base, pair.Quote)
}

func (o *Okex) getURL(pair Pair) string {
	return fmt.Sprintf(okexURL, o.localPairName(pair))
}

func (o *Okex) Fetch(pairs []Pair) []FetchResult {
	return callSinglePairOrigin(o, pairs)
}

func (o *Okex) callOne(pair Pair) (*Tick, error) {
	var err error
	req := &query.HTTPRequest{
		URL: o.getURL(pair),
	}

	// make query
	res := o.Pool.Query(req)
	if res == nil {
		return nil, errEmptyOriginResponse
	}
	if res.Error != nil {
		return nil, res.Error
	}
	// parsing JSON
	var resp okexResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse okex response: %w", err)
	}
	// parsing price from string
	price, err := strconv.ParseFloat(resp.Last, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse price from okex origin %s", res.Body)
	}
	// parsing ask price from string
	ask, err := strconv.ParseFloat(resp.Ask, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ask from okex origin %s", res.Body)
	}
	// parsing bid price from string
	bid, err := strconv.ParseFloat(resp.Bid, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bid from okex origin %s", res.Body)
	}
	// parsing volume from string
	volume, err := strconv.ParseFloat(resp.BaseVolume24H, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse volume from okex origin %s", res.Body)
	}
	// building Tick
	return &Tick{
		Pair:      pair,
		Timestamp: resp.Timestamp,
		Price:     price,
		Ask:       ask,
		Bid:       bid,
		Volume24h: volume,
	}, nil
}
