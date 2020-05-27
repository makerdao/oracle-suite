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
	"strconv"
	"strings"
)

// Huobi URL
const huobiURL = "https://api.huobi.pro/market/detail/merged?symbol=%s"

type huobiResponse struct {
	Status    string `json:"status"`
	Volume    string `json:"vol"`
	Timestamp int64  `json:"ts"`
	Tick      struct {
		Bid []string
	}
}

// Huobi exchange handler
type Huobi struct{}

// Call implementation
func (b *Huobi) Call(pool query.WorkerPool, pp *model.PotentialPricePoint) (*model.PricePoint, error) {
	if pool == nil {
		return nil, errNoPoolPassed
	}
	err := model.ValidatePotentialPricePoint(pp)
	if err != nil {
		return nil, err
	}

	pair := strings.ToLower(pp.Pair.Base + pp.Pair.Quote)
	req := &query.HTTPRequest{
		URL: fmt.Sprintf(huobiURL, pair),
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
	var resp huobiResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to pargse huobi response: %s", err)
	}
	if resp.Status == "error" {
		return nil, fmt.Errorf("wrong response from huobi exchange %s", res.Body)
	}
	if len(resp.Tick.Bid) < 1 {
		return nil, fmt.Errorf("wrong bid response from huobi exchange %s", res.Body)
	}
	// Parsing price from string
	price, err := strconv.ParseFloat(resp.Tick.Bid[0], 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse price from huobi exchange %s", res.Body)
	}
	// Parsing volume from string
	volume, err := strconv.ParseFloat(resp.Volume, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse volume from huobi exchange %s", res.Body)
	}
	// building PricePoint
	return &model.PricePoint{
		Exchange:  pp.Exchange,
		Pair:      pp.Pair,
		Price:     price,
		Volume:    volume,
		Timestamp: resp.Timestamp / 1000,
	}, nil
}
