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
	"time"
)

// Kraken URL
const krakenURL = "https://api.kraken.com/0/public/Ticker?pair=%s"

type krakenPairResponse struct {
	Price  []string `json:"c"`
	Volume []string `json:"v"`
}

type krakenResponse struct {
	Errors []string `json:"error"`
	Result map[string]*krakenPairResponse
}

// Kraken exchange handler
type Kraken struct{}

// Call implementation
func (b *Kraken) Call(pool query.WorkerPool, pp *model.PotentialPricePoint) (*model.PricePoint, error) {
	if pool == nil {
		return nil, errNoPoolPassed
	}
	err := model.ValidatePotentialPricePoint(pp)
	if err != nil {
		return nil, err
	}

	var pair string
	pair, ok := pp.Exchange.Config["pair"]
	if !ok || pair == "" {
		pair = strings.ToUpper(pp.Pair.Base + pp.Pair.Quote)
	}

	req := &query.HTTPRequest{
		URL: fmt.Sprintf(krakenURL, pair),
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
	var resp krakenResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to pargse kraken response: %s", err)
	}
	if len(resp.Errors) > 0 {
		return nil, fmt.Errorf("kraken API error: %s", strings.Join(resp.Errors, " "))
	}
	result, ok := resp.Result[pair]
	if !ok || result == nil {
		return nil, fmt.Errorf("wrong kraken exchange response. No resulting data %+v", resp)
	}
	if len(result.Price) == 0 || len(result.Volume) == 0 {
		return nil, fmt.Errorf("wrong kraken exchange response. No resulting pair %s data %+v", pair, result)
	}
	// Parsing price from string
	price, err := strconv.ParseFloat(result.Price[0], 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse price from kraken exchange %s", res.Body)
	}
	// Parsing volume from string
	volume, err := strconv.ParseFloat(result.Volume[0], 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse volume from kraken exchange %s", res.Body)
	}
	// building PricePoint
	return &model.PricePoint{
		Exchange:  pp.Exchange,
		Pair:      pp.Pair,
		Price:     price,
		Volume:    volume,
		Timestamp: time.Now().Unix(),
	}, nil
}
