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
	"strings"
)

// Upbit URL
const upbitURL = "https://api.upbit.com/v1/ticker?markets=%s"

type upbitResponse struct {
	Price     float64 `json:"trade_price"`
	Volume    float64 `json:"acc_trade_volume"`
	Timestamp int64   `json:"timestamp"`
}

// Upbit exchange handler
type Upbit struct{}

// Call implementation
func (b *Upbit) Call(pool query.WorkerPool, pp *model.PotentialPricePoint) (*model.PricePoint, error) {
	if pool == nil {
		return nil, errNoPoolPassed
	}
	err := model.ValidatePotentialPricePoint(pp)
	if err != nil {
		return nil, err
	}

	pair := fmt.Sprintf("%s-%s", strings.ToUpper(pp.Pair.Quote), strings.ToUpper(pp.Pair.Base))
	req := &query.HTTPRequest{
		URL: fmt.Sprintf(upbitURL, pair),
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
	var resp []upbitResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to pargse upbit response: %s", err)
	}
	if len(resp) < 1 {
		return nil, fmt.Errorf("wrong upbit response: %s", res.Body)
	}
	data := resp[0]
	// building PricePoint
	return &model.PricePoint{
		Exchange:  pp.Exchange,
		Pair:      pp.Pair,
		Price:     model.PriceFromFloat(data.Price),
		Volume:    model.PriceFromFloat(data.Volume),
		Timestamp: data.Timestamp / 1000,
	}, nil
}
