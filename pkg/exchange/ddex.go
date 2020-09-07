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
	"time"

	"github.com/makerdao/gofer/internal/query"
	"github.com/makerdao/gofer/pkg/model"
)

// Ddex URL
const ddexURL = "https://api.ddex.io/v4/markets/%s/orderbook?level=1"

type ddexResponse struct {
	Desc string `json:"desc"`
	Data struct {
		Orderbook struct {
			Bids []struct {
				Price  string `json:"price"`
				Amount string `json:"amount"`
			} `json:"bids"`
			Asks []struct {
				Price  string `json:"price"`
				Amount string `json:"amount"`
			} `json:"asks"`
		} `json:"orderbook"`
	} `json:"data"`
}

// Ddex exchange handler
type Ddex struct {
	Pool query.WorkerPool
}

func (d *Ddex) localPairName(pair *model.Pair) string {
	return fmt.Sprintf("%s-%s", pair.Base, pair.Quote)
}

func (d *Ddex) getURL(pp *model.PricePoint) string {
	return fmt.Sprintf(ddexURL, d.localPairName(pp.Pair))
}

func (d *Ddex) Fetch(pps []*model.PricePoint) {
	for _, pp := range pps {
		d.callOne(pp)
	}
}

func (d *Ddex) callOne(pp *model.PricePoint) {
	err := model.ValidatePricePoint(pp)
	if err != nil {
		pp.Error = err
		return
	}

	req := &query.HTTPRequest{
		URL: d.getURL(pp),
	}

	// make query
	res := d.Pool.Query(req)
	if res == nil {
		pp.Error = errEmptyExchangeResponse
		return
	}
	if res.Error != nil {
		pp.Error = res.Error
		return
	}
	// parsing JSON
	var resp ddexResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		pp.Error = fmt.Errorf("failed to parse ddex response: %w", err)
		return
	}
	// Check if response is successful
	if resp.Desc != "success" || len(resp.Data.Orderbook.Asks) != 1 || 1 != len(resp.Data.Orderbook.Bids) {
		pp.Error = fmt.Errorf("response returned from ddex exchange is invalid %s", res.Body)
		return
	}
	// Parsing ask from string
	ask, err := strconv.ParseFloat(resp.Data.Orderbook.Asks[0].Price, 64)
	if err != nil {
		pp.Error = fmt.Errorf("failed to parse ask from ddex exchange %s", res.Body)
		return
	}
	// Parsing bid from string
	bid, err := strconv.ParseFloat(resp.Data.Orderbook.Bids[0].Price, 64)
	if err != nil {
		pp.Error = fmt.Errorf("failed to parse bid from ddex exchange %s", res.Body)
		return
	}

	pp.Ask = ask
	pp.Bid = bid
	pp.Price = bid
	pp.Timestamp = time.Now().Unix()
}
