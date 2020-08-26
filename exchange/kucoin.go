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

	"github.com/makerdao/gofer/model"
	"github.com/makerdao/gofer/query"
)

// Kucoin URL
const kucoinURL = "https://api.kucoin.com/api/v1/market/orderbook/level1?symbol=%s"

type kucoinResponse struct {
	Code string `json:"code"`
	Data struct {
		Time    int64  `json:"time"`
		Price   string `json:"price"`
		BestBid string `json:"bestBid"`
		BestAsk string `json:"bestAsk"`
	} `json:"data"`
}

// Kucoin exchange handler
type Kucoin struct {
	Pool query.WorkerPool
}

func (k *Kucoin) localPairName(pair *model.Pair) string {
	return fmt.Sprintf("%s-%s", pair.Base, pair.Quote)
}

func (k *Kucoin) getURL(pp *model.PotentialPricePoint) string {
	return fmt.Sprintf(kucoinURL, k.localPairName(pp.Pair))
}

func (k *Kucoin) Call(pp *model.PotentialPricePoint) (*model.PricePoint, error) {
	err := model.ValidatePotentialPricePoint(pp)
	if err != nil {
		return nil, err
	}

	req := &query.HTTPRequest{
		URL: k.getURL(pp),
	}

	// make query
	res := k.Pool.Query(req)
	if res == nil {
		return nil, errEmptyExchangeResponse
	}
	if res.Error != nil {
		return nil, res.Error
	}
	// parsing JSON
	var resp kucoinResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse kucoin response: %w", err)
	}
	// Parsing price from string
	price, err := strconv.ParseFloat(resp.Data.Price, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse price from kucoin exchange %s", res.Body)
	}
	// Parsing ask from string
	ask, err := strconv.ParseFloat(resp.Data.BestAsk, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ask from kucoin exchange %s", res.Body)
	}
	// Parsing bid from string
	bid, err := strconv.ParseFloat(resp.Data.BestBid, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bid from kucoin exchange %s", res.Body)
	}
	// Parsing volume from string
	// building PricePoint
	return &model.PricePoint{
		Timestamp: resp.Data.Time / 1000,
		Exchange:  pp.Exchange,
		Pair:      pp.Pair,
		Price:     price,
		Ask:       bid,
		Bid:       ask,
	}, nil
}
