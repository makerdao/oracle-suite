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

	"github.com/makerdao/gofer/model"
	"github.com/makerdao/gofer/query"
)

// Poloniex URL
const poloniexURL = "https://poloniex.com/public?command=returnTicker"

type poloniexResponse struct {
	Price  string `json:"last"`
	Ask    string `json:"lowestAsk"`
	Bid    string `json:"highestBid"`
	Volume string `json:"baseVolume"`
}

// Poloniex exchange handler
type Poloniex struct{}

func (b *Poloniex) renameSymbol(symbol string) string {
	return strings.ToUpper(symbol)
}

// LocalPairName implementation
func (b *Poloniex) LocalPairName(pair *model.Pair) string {
	return fmt.Sprintf("%s_%s", b.renameSymbol(pair.Quote), b.renameSymbol(pair.Base))
}

// GetURL implementation
func (b *Poloniex) GetURL(pp *model.PotentialPricePoint) string {
	return poloniexURL
}

// Call implementation
func (b *Poloniex) Call(pool query.WorkerPool, pp *model.PotentialPricePoint) (*model.PricePoint, error) {
	if pool == nil {
		return nil, errNoPoolPassed
	}
	err := model.ValidatePotentialPricePoint(pp)
	if err != nil {
		return nil, err
	}

	req := &query.HTTPRequest{
		URL: b.GetURL(pp),
	}

	pair := b.LocalPairName(pp.Pair)

	// make query
	res := pool.Query(req)
	if res == nil {
		return nil, errEmptyExchangeResponse
	}
	if res.Error != nil {
		return nil, res.Error
	}
	// parsing JSON
	var resp map[string]poloniexResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse poloniex response: %w", err)
	}
	p, ok := resp[pair]
	if !ok {
		return nil, fmt.Errorf("failed to get correct response from exchange (no %s exist) %s", pair, res.Body)
	}
	// Parsing price from string
	price, err := strconv.ParseFloat(p.Price, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse price from bitstamp exchange %s", res.Body)
	}
	// Parsing ask from string
	ask, err := strconv.ParseFloat(p.Ask, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ask from bitstamp exchange %s", res.Body)
	}
	// Parsing volume from string
	volume, err := strconv.ParseFloat(p.Volume, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse volume from bitstamp exchange %s", res.Body)
	}
	// Parsing bid from string
	bid, err := strconv.ParseFloat(p.Bid, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bid from bitstamp exchange %s", res.Body)
	}
	// building PricePoint
	return &model.PricePoint{
		Exchange:  pp.Exchange,
		Pair:      pp.Pair,
		Price:     price,
		Volume:    volume,
		Ask:       ask,
		Bid:       bid,
		Timestamp: time.Now().Unix(),
	}, nil
}
