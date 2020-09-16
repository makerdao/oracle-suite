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

// Binance URL
const binanceURL = "https://www.binance.com/api/v3/ticker/price?symbol=%s"

type binanceResponse struct {
	Price string `json:"price"`
}

// Binance origin handler
type Binance struct {
	Pool query.WorkerPool
}

func (b *Binance) renameSymbol(symbol string) string {
	return strings.ToUpper(symbol)
}

func (b *Binance) localPairName(pair Pair) string {
	return b.renameSymbol(pair.Base) + b.renameSymbol(pair.Quote)
}

func (b *Binance) getURL(pair Pair) string {
	return fmt.Sprintf(binanceURL, b.localPairName(pair))
}

func (b *Binance) Fetch(pairs []Pair) []FetchResult {
	return callSinglePairOrigin(b, pairs)
}

func (b *Binance) callOne(pair Pair) (*Tick, error) {
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
	var resp binanceResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse binance response: %w", err)
	}
	// Parsing price from string
	price, err := strconv.ParseFloat(resp.Price, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse price from binance origin %s", res.Body)
	}
	// building Tick
	return &Tick{
		Pair:      pair,
		Price:     price,
		Timestamp: time.Now(),
	}, nil
}
