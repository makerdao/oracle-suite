//  Copyright (C) 2020  Maker Foundation
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
	"makerdao/gofer/model"
	"makerdao/gofer/query"
	"strconv"
	"strings"
	"time"
)

// Coinbase URL
const coinbaseProURL = "https://api.pro.coinbase.com/products//%s/ticker"

type coinbaseProResponse struct {
	Price  string `json:"price"`
	Ask    string `json:"ask"`
	Bid    string `json:"bid"`
	Volume string `json:"volume"`
}

// Coinbase exchange handler
type CoinbasePro struct{}

// Call implementation
func (b *CoinbasePro) Call(pool query.WorkerPool, pp *model.PotentialPricePoint) (*model.PricePoint, error) {
	if pool == nil {
		return nil, errNoPoolPassed
	}
	err := model.ValidatePotentialPricePoint(pp)
	if err != nil {
		return nil, err
	}

	pair := fmt.Sprintf("%s-%s", strings.ToUpper(pp.Pair.Base), strings.ToUpper(pp.Pair.Quote))
	req := &query.HTTPRequest{
		URL: fmt.Sprintf(coinbaseProURL, pair),
	}

	// make query
	res := pool.Query(req)
	if res == nil {
		return nil, errEmptyExchangeResponse
	}
	if res.Error != nil {
		return nil, res.Error
	}
	// parsing JSON
	var resp coinbaseProResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to pargse coinbasepro response: %s", err)
	}
	// Parsing price from string
	price, err := strconv.ParseFloat(resp.Price, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse price from coinbasepro exchange %s", res.Body)
	}
	// Parsing ask from string
	ask, err := strconv.ParseFloat(resp.Ask, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ask from coinbasepro exchange %s", res.Body)
	}
	// Parsing volume from string
	volume, err := strconv.ParseFloat(resp.Volume, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse volume from coinbasepro exchange %s", res.Body)
	}
	// Parsing bid from string
	bid, err := strconv.ParseFloat(resp.Bid, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bid from coinbasepro exchange %s", res.Body)
	}
	// building PricePoint
	return &model.PricePoint{
		Exchange:  pp.Exchange,
		Pair:      pp.Pair,
		Price:     model.PriceFromFloat(price),
		Volume:    model.PriceFromFloat(volume),
		Ask:       model.PriceFromFloat(ask),
		Bid:       model.PriceFromFloat(bid),
		Timestamp: time.Now().Unix(),
	}, nil
}
