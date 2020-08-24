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
	"time"

	"github.com/makerdao/gofer/model"
	"github.com/makerdao/gofer/query"
)

// Exchange URL
const ftxURL = "https://ftx.com/api/markets/%s"

type ftxResponse struct {
	Result struct {
		Ask    float64 `json:"ask"`
		Bid    float64 `json:"bid"`
		Price  float64 `json:"last"`
		Volume float64 `json:"quoteVolume24h"`
		Name   string  `json:"name"`
	}
	Success bool `json:"success"`
}

// Exchange handler
type Ftx struct{}

func (b *Ftx) localPairName(pair *model.Pair) string {
	return fmt.Sprintf("%s/%s", pair.Base, pair.Quote)
}

func (b *Ftx) getURL(pp *model.PotentialPricePoint) string {
	return fmt.Sprintf(ftxURL, b.localPairName(pp.Pair))
}

// Call implementation
func (b *Ftx) Call(pool query.WorkerPool, pp *model.PotentialPricePoint) (*model.PricePoint, error) {
	if pool == nil {
		return nil, errNoPoolPassed
	}
	err := model.ValidatePotentialPricePoint(pp)
	if err != nil {
		return nil, err
	}

	req := &query.HTTPRequest{
		URL: b.getURL(pp),
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
	var resp ftxResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ftx response: %w", err)
	}

	if !resp.Success || resp.Result.Name != b.localPairName(pp.Pair) {
		return nil, fmt.Errorf("failed to get correct response from ftx: %s", res.Body)
	}

	// building PricePoint
	return &model.PricePoint{
		Exchange:  pp.Exchange,
		Pair:      pp.Pair,
		Price:     resp.Result.Price,
		Ask:       resp.Result.Ask,
		Bid:       resp.Result.Bid,
		Volume:    resp.Result.Volume,
		Timestamp: time.Now().Unix(),
	}, nil
}
