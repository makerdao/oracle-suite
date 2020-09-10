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
	"strconv"
	"strings"
	"time"

	"github.com/makerdao/gofer/internal/query"
)

// Hitbtc URL
const hitbtcURL = "https://api.hitbtc.com/api/2/public/ticker/%s"

type hitbtcResponse struct {
	Ask       string    `json:"ask"`
	Volume    string    `json:"volume"`
	Price     string    `json:"last"`
	Bid       string    `json:"bid"`
	Timestamp time.Time `json:"timestamp"`
}

// Hitbtc exchange handler
type Hitbtc struct {
	Pool query.WorkerPool
}

func (h *Hitbtc) localPairName(pair Pair) string {
	return strings.ToUpper(pair.Base + pair.Quote)
}

func (h *Hitbtc) getURL(pair Pair) string {
	return fmt.Sprintf(hitbtcURL, h.localPairName(pair))
}

func (h *Hitbtc) Call(pairs []Pair) []CallResult {
	return callSinglePairExchange(h, pairs)
}

func (h *Hitbtc) callOne(pair Pair) (*Tick, error) {
	var err error
	req := &query.HTTPRequest{
		URL: h.getURL(pair),
	}

	// make query
	res := h.Pool.Query(req)
	if res == nil {
		return nil, errEmptyExchangeResponse
	}
	if res.Error != nil {
		return nil, res.Error
	}
	// parsing JSON
	var resp hitbtcResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse hitbtc response: %w", err)
	}
	// Parsing price from string
	price, err := strconv.ParseFloat(resp.Price, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse price from hitbtc exchange %s", res.Body)
	}
	// Parsing ask from string
	ask, err := strconv.ParseFloat(resp.Ask, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ask from hitbtc exchange %s", res.Body)
	}
	// Parsing volume from string
	volume, err := strconv.ParseFloat(resp.Volume, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse volume from hitbtc exchange %s", res.Body)
	}
	// Parsing bid from string
	bid, err := strconv.ParseFloat(resp.Bid, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bid from hitbtc exchange %s", res.Body)
	}
	// building Tick
	return &Tick{
		Pair:      pair,
		Price:     price,
		Volume24h: volume,
		Ask:       ask,
		Bid:       bid,
		Timestamp: resp.Timestamp,
	}, nil
}
