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

func (d *Ddex) getURL(pp *model.PotentialPricePoint) string {
	return fmt.Sprintf(ddexURL, d.localPairName(pp.Pair))
}

func (d *Ddex) Call(ppps []*model.PotentialPricePoint) []CallResult {
	return callSinglePairExchange(d, ppps)
}

func (d *Ddex) callOne(pp *model.PotentialPricePoint) (*model.PricePoint, error) {
	err := model.ValidatePotentialPricePoint(pp)
	if err != nil {
		return nil, err
	}

	req := &query.HTTPRequest{
		URL: d.getURL(pp),
	}

	// make query
	res := d.Pool.Query(req)
	if res == nil {
		return nil, errEmptyExchangeResponse
	}
	if res.Error != nil {
		return nil, res.Error
	}
	// parsing JSON
	var resp ddexResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ddex response: %w", err)
	}
	// Check if response is successful
	if resp.Desc != "success" || len(resp.Data.Orderbook.Asks) != 1 || 1 != len(resp.Data.Orderbook.Bids) {
		return nil, fmt.Errorf("response returned from ddex exchange is invalid %s", res.Body)
	}
	// Parsing ask from string
	ask, err := strconv.ParseFloat(resp.Data.Orderbook.Asks[0].Price, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ask from ddex exchange %s", res.Body)
	}
	// Parsing bid from string
	bid, err := strconv.ParseFloat(resp.Data.Orderbook.Bids[0].Price, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bid from ddex exchange %s", res.Body)
	}
	// building PricePoint
	return &model.PricePoint{
		Exchange:  pp.Exchange,
		Pair:      pp.Pair,
		Ask:       ask,
		Bid:       bid,
		Price:     bid,
		Timestamp: time.Now().Unix(),
	}, nil
}
