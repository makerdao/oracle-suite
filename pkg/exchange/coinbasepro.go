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
	"github.com/makerdao/gofer/pkg/model"
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
type CoinbasePro struct {
	Pool query.WorkerPool
}

func (c *CoinbasePro) localPairName(pair *model.Pair) string {
	return fmt.Sprintf("%s-%s", strings.ToUpper(pair.Base), strings.ToUpper(pair.Quote))
}

func (c *CoinbasePro) getURL(pp *model.PricePoint) string {
	return fmt.Sprintf(coinbaseProURL, c.localPairName(pp.Pair))
}

func (c *CoinbasePro) Fetch(pps []*model.PricePoint) {
	for _, pp := range pps {
		c.callOne(pp)
	}
}

func (c *CoinbasePro) callOne(pp *model.PricePoint) {
	err := model.ValidatePricePoint(pp)
	if err != nil {
		pp.Error = err
		return
	}

	req := &query.HTTPRequest{
		URL: c.getURL(pp),
	}

	// make query
	res := c.Pool.Query(req)
	if res == nil {
		pp.Error = errEmptyExchangeResponse
		return
	}
	if res.Error != nil {
		pp.Error = res.Error
		return
	}
	// parsing JSON
	var resp coinbaseProResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		pp.Error = fmt.Errorf("failed to parse coinbasepro response: %w", err)
		return
	}
	// Parsing price from string
	price, err := strconv.ParseFloat(resp.Price, 64)
	if err != nil {
		pp.Error = fmt.Errorf("failed to parse price from coinbasepro exchange %s", res.Body)
		return
	}
	// Parsing ask from string
	ask, err := strconv.ParseFloat(resp.Ask, 64)
	if err != nil {
		pp.Error = fmt.Errorf("failed to parse ask from coinbasepro exchange %s", res.Body)
		return
	}
	// Parsing volume from string
	volume, err := strconv.ParseFloat(resp.Volume, 64)
	if err != nil {
		pp.Error = fmt.Errorf("failed to parse volume from coinbasepro exchange %s", res.Body)
		return
	}
	// Parsing bid from string
	bid, err := strconv.ParseFloat(resp.Bid, 64)
	if err != nil {
		pp.Error = fmt.Errorf("failed to parse bid from coinbasepro exchange %s", res.Body)
		return
	}

	pp.Price = price
	pp.Volume = volume
	pp.Ask = ask
	pp.Bid = bid
	pp.Timestamp = time.Now().Unix()
}
