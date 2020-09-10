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
	"strings"
	"time"

	"github.com/makerdao/gofer/internal/query"
)

// Huobi URL
const huobiURL = "https://api.huobi.pro/market/detail/merged?symbol=%s"

type huobiResponse struct {
	Status    string  `json:"status"`
	Volume    float64 `json:"vol"`
	Timestamp int64   `json:"ts"`
	Tick      struct {
		Bid []float64
	}
}

// Huobi exchange handler
type Huobi struct {
	Pool query.WorkerPool
}

func (h *Huobi) localPairName(pair Pair) string {
	return strings.ToLower(pair.Base + pair.Quote)
}

func (h *Huobi) getURL(pair Pair) string {
	return fmt.Sprintf(huobiURL, h.localPairName(pair))
}

func (h *Huobi) Call(pairs []Pair) []CallResult {
	return callSinglePairExchange(h, pairs)
}

func (h *Huobi) callOne(pair Pair) (*Tick, error) {
	var err error
	req := &query.HTTPRequest{
		URL: h.getURL(pair),
	}

	res := h.Pool.Query(req)
	if res == nil {
		return nil, errEmptyExchangeResponse
	}
	if res.Error != nil {
		return nil, res.Error
	}

	var resp huobiResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse huobi response: %w", err)
	}
	if resp.Status == "error" {
		return nil, fmt.Errorf("wrong response from huobi exchange %s", res.Body)
	}
	if len(resp.Tick.Bid) < 1 {
		return nil, fmt.Errorf("wrong bid response from huobi exchange %s", res.Body)
	}

	return &Tick{
		Pair:      pair,
		Price:     resp.Tick.Bid[0],
		Volume24h: resp.Volume,
		Timestamp: time.Unix(resp.Timestamp/1000, 0),
	}, nil
}
