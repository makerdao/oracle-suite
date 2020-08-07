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

	"github.com/makerdao/gofer/model"
	"github.com/makerdao/gofer/query"
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

func (k *Kraken) getPair(pp *model.PotentialPricePoint) string {
	pair, ok := pp.Exchange.Config["pair"]
	if !ok || pair == "" {
		pair = k.LocalPairName(pp.Pair)
	}
	return pair
}

func (k *Kraken) getSymbol(symbol string) string {
	switch strings.ToUpper(symbol) {
	case "BTC":
		return "XBT"
	default:
		return strings.ToUpper(symbol)
	}
}

// LocalPairName implementation
func (k *Kraken) LocalPairName(pair *model.Pair) string {
	if pair.Base == "USDT" {
		return strings.ToUpper(fmt.Sprintf("%sZ%s", k.getSymbol(pair.Base), k.getSymbol(pair.Quote)))
	}
	return strings.ToUpper(fmt.Sprintf("X%sZ%s", k.getSymbol(pair.Base), k.getSymbol(pair.Quote)))
}

// GetURL implementation
func (k *Kraken) GetURL(pp *model.PotentialPricePoint) string {
	return fmt.Sprintf(krakenURL, k.getPair(pp))
}

// Call implementation
func (k *Kraken) Call(pool query.WorkerPool, pp *model.PotentialPricePoint) (*model.PricePoint, error) {
	if pool == nil {
		return nil, errNoPoolPassed
	}
	err := model.ValidatePotentialPricePoint(pp)
	if err != nil {
		return nil, err
	}

	req := &query.HTTPRequest{
		URL: k.GetURL(pp),
	}
	pair := k.getPair(pp)

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
		return nil, fmt.Errorf("failed to parse kraken response: %w", err)
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
