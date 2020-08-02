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
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/makerdao/gofer/model"
	"github.com/makerdao/gofer/query"
)

// Uniswap URL
const uniswapURL = "https://api.thegraph.com/subgraphs/name/uniswap/uniswap-v2"

type uniswapPairResponse struct {
	Price string `json:"token1Price"`
}

type uniswapResponse struct {
	Data struct {
		Pairs []*uniswapPairResponse
	}
}

// Uniswap exchange handler
type Uniswap struct{}

// LocalPairName implementation
func (k *Uniswap) LocalPairName(pair *model.Pair) string {
	switch *pair {
	case *model.NewPair("COMP", "ETH"):
		return "0xcffdded873554f362ac02f8fb1f02e5ada10516f"
	case *model.NewPair("LRC", "ETH"):
		return "0x8878df9e1a7c87dcbf6d3999d997f262c05d8c70"
	default:
		return pair.String()
	}
}

// GetURL implementation
func (k *Uniswap) GetURL(_ *model.PotentialPricePoint) string {
	return uniswapURL
}

// Call implementation
func (k *Uniswap) Call(pool query.WorkerPool, pp *model.PotentialPricePoint) (*model.PricePoint, error) {
	if pool == nil {
		return nil, errNoPoolPassed
	}
	err := model.ValidatePotentialPricePoint(pp)
	if err != nil {
		return nil, err
	}

	pair := k.LocalPairName(pp.Pair)
	body := fmt.Sprintf(`{"query":"query($id:String){pairs(where:{id:$id}){token1Price}}","variables":{"id":"%s"}}`, pair)

	req := &query.HTTPRequest{
		URL:    k.GetURL(pp),
		Method: "POST",
		Body:   bytes.NewBuffer([]byte(body)),
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
	var resp uniswapResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse uniswap response: %w", err)
	}
	if len(resp.Data.Pairs) == 0 {
		return nil, fmt.Errorf("failed to parse uniswap response: no pairs %s", res.Body)
	}
	if resp.Data.Pairs[0].Price == "" {
		return nil, fmt.Errorf("failed to parse uniswap price: %s", res.Body)
	}
	// Parsing price from string
	price, err := strconv.ParseFloat(resp.Data.Pairs[0].Price, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse price from uniswap exchange %s", res.Body)
	}
	// building PricePoint
	return &model.PricePoint{
		Exchange:  pp.Exchange,
		Pair:      pp.Pair,
		Price:     price,
		Timestamp: time.Now().Unix(),
	}, nil
}
