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
	"strings"
	"time"

	"github.com/makerdao/oracle-suite/internal/query"
)

const uniswapV3URL = "https://api.thegraph.com/subgraphs/name/ianlapham/uniswap-v3-alt"

type uniswapV3Response struct {
	Data struct {
		Pools []uniswapV3PairResponse
	}
}

type uniswapV3TokenResponse struct {
	Symbol string `json:"symbol"`
}

type uniswapV3PairResponse struct {
	ID      string                 `json:"id"`
	Price0  stringAsFloat64        `json:"token0Price"`
	Price1  stringAsFloat64        `json:"token1Price"`
	Volume0 stringAsFloat64        `json:"volumeToken0"`
	Volume1 stringAsFloat64        `json:"volumeToken1"`
	Token0  uniswapV3TokenResponse `json:"token0"`
	Token1  uniswapV3TokenResponse `json:"token1"`
}

type UniswapV3 struct {
	Pool query.WorkerPool
}

func (u *UniswapV3) pairsToContractAddress(pair Pair) string {
	// We're checking for reverse pairs because the same contract is used to
	// trade in both directions.
	match := func(a, b Pair) bool {
		if a.Quote == b.Quote && a.Base == b.Base {
			return true
		}

		if a.Quote == b.Base && a.Base == b.Quote {
			return true
		}

		return false
	}

	p := Pair{Base: u.renameSymbol(pair.Base), Quote: u.renameSymbol(pair.Quote)}

	switch {
	case match(p, Pair{Base: "COMP", Quote: "WETH"}):
		return "0xea4ba4ce14fdd287f380b55419b1c5b6c3f22ab6"
	case match(p, Pair{Base: "CRV", Quote: "WETH"}):
		return "0x04916039b1f59d9745bf6e0a21f191d1e0a84287"
	case match(p, Pair{Base: "WETH", Quote: "USDC"}):
		return "0x8ad599c3a0ff1de082011efddc58f1908eb6e6d8"
	case match(p, Pair{Base: "WETH", Quote: "KNC"}):
		return "0x76838fd2f22bdc1d3e96069971e65653173edb2a"
	case match(p, Pair{Base: "LRC", Quote: "WETH"}):
		return "0xe1d92f1de49caec73514f696fea2a7d5441498e5"
	case match(p, Pair{Base: "UNI", Quote: "WETH"}):
		return "0x1d42064fc4beb5f8aaf85f4617ae8b3b5b8bd801"
	case match(p, Pair{Base: "WNXM", Quote: "WETH"}):
		return "0x058d79a4c6eb5b11d0248993ffa1faa168ddd3c0"
	case match(p, Pair{Base: "YFI", Quote: "WETH"}):
		return "0x04916039b1f59d9745bf6e0a21f191d1e0a84287"
	}
	return pair.String()
}

// TODO: We should find better solution for this.
func (u *UniswapV3) renameSymbol(symbol string) string {
	switch symbol {
	case "ETH":
		return "WETH"
	case "BTC":
		return "WBTC"
	case "USD":
		return "USDC"
	}
	return symbol
}

func (u *UniswapV3) Fetch(pairs []Pair) []FetchResult {
	return callSinglePairOrigin(u, pairs)
}

//nolint:dupl
func (u *UniswapV3) callOne(pair Pair) (*Price, error) {
	var err error

	pairsJSON, _ := json.Marshal(u.pairsToContractAddress(pair))
	gql := `
		query($id:String) {
			pools(where:{id: $id}) {
				id
				token0Price
				token1Price
				volumeToken0
				volumeToken1
				token0 { symbol }
				token1 { symbol }
			}
		}
	`
	body := fmt.Sprintf(
		`{"query":"%s","variables":{"id":%s}}`,
		strings.ReplaceAll(strings.ReplaceAll(gql, "\n", " "), "\t", ""),
		pairsJSON,
	)

	req := &query.HTTPRequest{
		URL:    uniswapV3URL,
		Method: "POST",
		Body:   bytes.NewBuffer([]byte(body)),
	}

	// make query
	res := u.Pool.Query(req)
	if res == nil {
		return nil, ErrEmptyOriginResponse
	}
	if res.Error != nil {
		return nil, res.Error
	}

	// parse JSON
	var resp uniswapV3Response
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse UniswapV3 response: %w", err)
	}

	// convert response from a slice to a map
	respMap := map[string]uniswapV3PairResponse{}
	for _, pairResp := range resp.Data.Pools {
		respMap[pairResp.Token0.Symbol+"/"+pairResp.Token1.Symbol] = pairResp
	}

	b := u.renameSymbol(pair.Base)
	q := u.renameSymbol(pair.Quote)

	pair0 := b + "/" + q
	pair1 := q + "/" + b

	if r, ok := respMap[pair0]; ok {
		return &Price{
			Pair:      pair,
			Price:     r.Price1.val(),
			Bid:       r.Price1.val(),
			Ask:       r.Price1.val(),
			Volume24h: r.Volume0.val(),
			Timestamp: time.Now(),
		}, nil
	} else if r, ok := respMap[pair1]; ok {
		return &Price{
			Pair:      pair,
			Price:     r.Price0.val(),
			Bid:       r.Price0.val(),
			Ask:       r.Price0.val(),
			Volume24h: r.Volume1.val(),
			Timestamp: time.Now(),
		}, nil
	}
	return nil, ErrMissingResponseForPair
}
