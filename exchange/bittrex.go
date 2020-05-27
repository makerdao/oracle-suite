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
	"github.com/makerdao/gofer/model"
	"github.com/makerdao/gofer/query"
	"strings"
	"time"
)

// BitTrex URL
const bittrexURL = "https://api.bittrex.com/api/v1.1/public/getticker?market=%s"

type bittrexResponse struct {
	Success bool `json:"success"`
	Result  struct {
		Ask  float64 `json:"Ask"`
		Bid  float64 `json:"Bid"`
		Last float64 `json:"Last"`
	} `json:"result"`
}

// BitTrex exchange handler
type BitTrex struct{}

// Call implementation
func (b *BitTrex) Call(pool query.WorkerPool, pp *model.PotentialPricePoint) (*model.PricePoint, error) {
	if pool == nil {
		return nil, errNoPoolPassed
	}
	err := model.ValidatePotentialPricePoint(pp)
	if err != nil {
		return nil, err
	}

	pair := fmt.Sprintf("%s-%s", strings.ToUpper(pp.Pair.Quote), strings.ToUpper(pp.Pair.Base))
	req := &query.HTTPRequest{
		URL: fmt.Sprintf(bittrexURL, pair),
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
	var resp bittrexResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to pargse bittrex response: %s", err)
	}
	if !resp.Success {
		return nil, fmt.Errorf("wrong response from bittrex %v", resp)
	}
	// building PricePoint
	return &model.PricePoint{
		Exchange:  pp.Exchange,
		Pair:      pp.Pair,
		Price:     resp.Result.Last,
		Ask:       resp.Result.Ask,
		Bid:       resp.Result.Bid,
		Timestamp: time.Now().Unix(),
	}, nil
}
