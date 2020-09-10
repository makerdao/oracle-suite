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
)

// Exchange URL
const cryptoCompareURL = "https://min-api.cryptocompare.com/data/price?fsym=%s&tsyms=%s"

type cryptoCompareResponse map[string]float64

// Exchange handler
type CryptoCompare struct {
	Pool query.WorkerPool
}

func (c *CryptoCompare) getURL(pp Pair) string {
	return fmt.Sprintf(cryptoCompareURL, pp.Base, pp.Quote)
}

func (c *CryptoCompare) Call(ppps []Pair) []CallResult {
	return callSinglePairExchange(c, ppps)
}

func (c *CryptoCompare) callOne(pp Pair) (*Tick, error) {
	var err error
	req := &query.HTTPRequest{
		URL: c.getURL(pp),
	}

	// make query
	res := c.Pool.Query(req)
	if res == nil {
		return nil, errEmptyExchangeResponse
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

	price, ok := resp[pp.Quote]
	if !ok {
		return nil, fmt.Errorf("failed to get price for %s: %s", pp.Quote, res.Body)
	}

	// building Tick
	return &Tick{
		Pair:      pp,
		Price:     price,
		Timestamp: time.Now(),
	}, nil
}
