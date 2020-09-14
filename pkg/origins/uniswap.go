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

package origins

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/makerdao/gofer/internal/query"
)

// Uniswap URL
const uniswapURL = "https://api.thegraph.com/subgraphs/name/uniswap/uniswap-v2"

type uniswapPairResponse struct {
	Price0 string `json:"token0Price"`
	Price  string `json:"token1Price"`
}

type uniswapResponse struct {
	Data struct {
		Pairs []*uniswapPairResponse
	}
}

func getPriceByPair(pair Pair, res *uniswapPairResponse) string {
	p := Pair{Base: "KNC", Quote: "ETH"}
	if pair == p {
		return res.Price0
	}
	return res.Price
}

// Uniswap origin handler
type Uniswap struct {
	Pool query.WorkerPool
}

func (u *Uniswap) localPairName(pair Pair) string {
	switch pair {
	case Pair{Base: "COMP", Quote: "ETH"}:
		return "0xcffdded873554f362ac02f8fb1f02e5ada10516f"
	case Pair{Base: "LRC", Quote: "ETH"}:
		return "0x8878df9e1a7c87dcbf6d3999d997f262c05d8c70"
	case Pair{Base: "KNC", Quote: "ETH"}:
		return "0xf49c43ae0faf37217bdcb00df478cf793edd6687"
	default:
		return pair.String()
	}
}

func (u *Uniswap) getURL(_ Pair) string {
	return uniswapURL
}

func (u *Uniswap) Fetch(pairs []Pair) []FetchResult {
	return callSinglePairOrigin(u, pairs)
}

func (u *Uniswap) callOne(pair Pair) (*Tick, error) {
	var err error
	pairName := u.localPairName(pair)
	body := fmt.Sprintf(
		`{"query":"query($id:String){pairs(where:{id:$id}){token0Price token1Price}}","variables":{"id":"%s"}}`,
		pairName,
	)

	req := &query.HTTPRequest{
		URL:    u.getURL(pair),
		Method: "POST",
		Body:   bytes.NewBuffer([]byte(body)),
	}

	// make query
	res := u.Pool.Query(req)
	if res == nil {
		return nil, errEmptyOriginResponse
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
	// Due to API for some pairs like `KNC/ETH` we have to take `token0Price` field rather than `token1Price`
	priceStr := getPriceByPair(pair, resp.Data.Pairs[0])
	if priceStr == "" {
		return nil, fmt.Errorf("failed to parse uniswap price: %s", res.Body)
	}
	// Parsing price from string
	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse price from uniswap origin %s", res.Body)
	}
	// building Tick
	return &Tick{
		Pair:      pair,
		Price:     price,
		Timestamp: time.Now(),
	}, nil
}
