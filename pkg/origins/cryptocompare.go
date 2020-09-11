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

// Origin URL
const cryptoCompareURL = "https://min-api.cryptocompare.com/data/price?fsym=%s&tsyms=%s"

type cryptoCompareResponse map[string]float64

// Origin handler
type CryptoCompare struct {
	Pool query.WorkerPool
}

func (c *CryptoCompare) getURL(pair Pair) string {
	return fmt.Sprintf(cryptoCompareURL, pair.Base, pair.Quote)
}

func (c *CryptoCompare) Fetch(pairs []Pair) []FetchResult {
	return callSinglePairOrigin(c, pairs)
}

func (c *CryptoCompare) callOne(pair Pair) (*Tick, error) {
	var err error
	req := &query.HTTPRequest{
		URL: c.getURL(pair),
	}

	// make query
	res := c.Pool.Query(req)
	if res == nil {
		return nil, errEmptyOriginResponse
	}
	if res.Error != nil {
		return nil, res.Error
	}
	// parsing JSON
	var resp cryptoCompareResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CryptoCompare response: %w", err)
	}

	price, ok := resp[pair.Quote]
	if !ok {
		return nil, fmt.Errorf("failed to get price for %s: %s", pair.Quote, res.Body)
	}

	// building Tick
	return &Tick{
		Pair:      pair,
		Price:     price,
		Timestamp: time.Now(),
	}, nil
}
