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

const sushiswapURL = "https://api.thegraph.com/subgraphs/name/zippoxer/sushiswap-subgraph-fork"

type sushiswapResponse struct {
	Data struct {
		Pairs []sushiswapPairResponse
	}
}

type sushiswapTokenResponse struct {
	Symbol string `json:"symbol"`
}

type sushiswapPairResponse struct {
	ID      string                 `json:"id"`
	Price0  stringAsFloat64        `json:"token0Price"`
	Price1  stringAsFloat64        `json:"token1Price"`
	Volume0 stringAsFloat64        `json:"volumeToken0"`
	Volume1 stringAsFloat64        `json:"volumeToken1"`
	Token0  sushiswapTokenResponse `json:"token0"`
	Token1  sushiswapTokenResponse `json:"token1"`
}

type Sushiswap struct {
	Pool query.WorkerPool
}

func (s *Sushiswap) pairsToContractAddress(pair Pair) string {
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

	p := Pair{Base: s.renameSymbol(pair.Base), Quote: s.renameSymbol(pair.Quote)}

	switch {
	case match(p, Pair{Base: "SNX", Quote: "WETH"}):
		return "0xa1d7b2d891e3a1f9ef4bbc5be20630c2feb1c470"
	case match(p, Pair{Base: "CRV", Quote: "WETH"}):
		return "0x58dc5a51fe44589beb22e8ce67720b5bc5378009"
	}

	return pair.String()
}

// TODO: We should find better solution for this.
func (s *Sushiswap) renameSymbol(symbol string) string {
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

func (s *Sushiswap) Fetch(pairs []Pair) []FetchResult {
	return callSinglePairOrigin(s, pairs)
}

func (s *Sushiswap) callOne(pair Pair) (*Price, error) {
	var err error

	pairsJSON, _ := json.Marshal(s.pairsToContractAddress(pair))
	gql := `
		query($id:String) {
			pairs(where:{id: $id}) {
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
		URL:    sushiswapURL,
		Method: "POST",
		Body:   bytes.NewBuffer([]byte(body)),
	}

	// make query
	res := s.Pool.Query(req)
	if res == nil {
		return nil, ErrEmptyOriginResponse
	}
	if res.Error != nil {
		return nil, res.Error
	}

	// parse JSON
	var resp sushiswapResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Sushiswap response: %w", err)
	}

	// convert response from a slice to a map
	respMap := map[string]sushiswapPairResponse{}
	for _, pairResp := range resp.Data.Pairs {
		respMap[pairResp.Token0.Symbol+"/"+pairResp.Token1.Symbol] = pairResp
	}

	b := s.renameSymbol(pair.Base)
	q := s.renameSymbol(pair.Quote)

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
