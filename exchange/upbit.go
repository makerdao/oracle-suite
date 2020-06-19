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

	"github.com/makerdao/gofer/model"
	"github.com/makerdao/gofer/query"
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

func (u *Upbit) renameSymbol(symbol string) string {
	return strings.ToUpper(symbol)
}

// LocalPairName implementation
func (u *Upbit) LocalPairName(pair *model.Pair) string {
	return fmt.Sprintf("%s-%s", u.renameSymbol(pair.Quote), u.renameSymbol(pair.Base))
}

// GetURL implementation
func (u *Upbit) GetURL(pp *model.PotentialPricePoint) string {
	return fmt.Sprintf(upbitURL, u.LocalPairName(pp.Pair))
}

// Call implementation
func (u *Upbit) Call(pool query.WorkerPool, pp *model.PotentialPricePoint) (*model.PricePoint, error) {
	if pool == nil {
		return nil, errNoPoolPassed
	}
	err := model.ValidatePotentialPricePoint(pp)
	if err != nil {
		return nil, err
	}

	req := &query.HTTPRequest{
		URL: u.GetURL(pp),
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
		return nil, fmt.Errorf("failed to parse upbit response: %w", err)
	}
	if len(resp) < 1 {
		return nil, fmt.Errorf("wrong upbit response: %s", res.Body)
	}
	data := resp[0]
	// building PricePoint
	return &model.PricePoint{
		Exchange:  pp.Exchange,
		Pair:      pp.Pair,
		Price:     data.Price,
		Volume:    data.Volume,
		Timestamp: data.Timestamp / 1000,
	}, nil
}
