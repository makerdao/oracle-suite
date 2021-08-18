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

const uniswapURL = "https://api.thegraph.com/subgraphs/name/uniswap/uniswap-v2"

type uniswapResponse struct {
	Data struct {
		Pairs []uniswapPairResponse
	}
}

type uniswapTokenResponse struct {
	Symbol string `json:"symbol"`
}

type uniswapPairResponse struct {
	ID      string               `json:"id"`
	Price0  stringAsFloat64      `json:"token0Price"`
	Price1  stringAsFloat64      `json:"token1Price"`
	Volume0 stringAsFloat64      `json:"volumeToken0"`
	Volume1 stringAsFloat64      `json:"volumeToken1"`
	Token0  uniswapTokenResponse `json:"token0"`
	Token1  uniswapTokenResponse `json:"token1"`
}

type Uniswap struct {
	WorkerPool        query.WorkerPool
	ContractAddresses ContractAddresses
}

//nolint:gocyclo
func (u *Uniswap) pairsToContractAddresses(pairs []Pair) ([]string, error) {
	var names []string

	for _, pair := range pairs {
		address, ok := u.ContractAddresses.ByPair(pair)
		if !ok {
			return names, fmt.Errorf("failed to find contract address for pair %s", pair.String())
		}
		names = append(names, address)
		//p := Pair{Base: u.renameSymbol(pair.Base), Quote: u.renameSymbol(pair.Quote)}
		//
		//switch {
		//case match(p, Pair{Base: "AAVE", Quote: "WETH"}):
		//	names = append(names, "0xdfc14d2af169b0d36c4eff567ada9b2e0cae044f")
		//case match(p, Pair{Base: "BAT", Quote: "WETH"}):
		//	names = append(names, "0xa70d458a4d9bc0e6571565faee18a48da5c0d593")
		//case match(p, Pair{Base: "SNX", Quote: "WETH"}):
		//	names = append(names, "0x43ae24960e5534731fc831386c07755a2dc33d47")
		//case match(p, Pair{Base: "COMP", Quote: "WETH"}):
		//	names = append(names, "0xcffdded873554f362ac02f8fb1f02e5ada10516f")
		//case match(p, Pair{Base: "WETH", Quote: "USDC"}):
		//	names = append(names, "0xb4e16d0168e52d35cacd2c6185b44281ec28c9dc")
		//case match(p, Pair{Base: "KNC", Quote: "WETH"}):
		//	names = append(names, "0xf49c43ae0faf37217bdcb00df478cf793edd6687")
		//case match(p, Pair{Base: "LEND", Quote: "WETH"}):
		//	names = append(names, "0xab3f9bf1d81ddb224a2014e98b238638824bcf20")
		//case match(p, Pair{Base: "LRC", Quote: "WETH"}):
		//	names = append(names, "0x8878df9e1a7c87dcbf6d3999d997f262c05d8c70")
		//case match(p, Pair{Base: "PAXG", Quote: "WETH"}):
		//	names = append(names, "0x9c4fe5ffd9a9fc5678cfbd93aa2d4fd684b67c4c")
		//case match(p, Pair{Base: "YFI", Quote: "WETH"}):
		//	names = append(names, "0x2fdbadf3c4d5a8666bc06645b8358ab803996e28")
		//}
	}
	return names, nil
}

//func (u *Uniswap) renameSymbol(symbol string) string {
//	switch symbol {
//	case "ETH":
//		return "WETH"
//	case "BTC":
//		return "WBTC"
//	case "USD":
//		return "USDC"
//	}
//	return symbol
//}

func (u Uniswap) Pool() query.WorkerPool {
	return u.WorkerPool
}

func (u Uniswap) PullPrices(pairs []Pair) []FetchResult {
	var err error

	contracts, err := u.pairsToContractAddresses(pairs)
	if err != nil {
		return fetchResultListWithErrors(pairs, err)
	}
	pairsJSON, _ := json.Marshal(contracts)
	gql := `
		query($ids:[String]) {
			pairs(where:{id_in:$ids}) {
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
		`{"query":"%s","variables":{"ids":%s}}`,
		strings.ReplaceAll(strings.ReplaceAll(gql, "\n", " "), "\t", ""),
		pairsJSON,
	)

	req := &query.HTTPRequest{
		URL:    uniswapURL,
		Method: "POST",
		Body:   bytes.NewBuffer([]byte(body)),
	}

	// make query
	res := u.WorkerPool.Query(req)
	if res == nil {
		return fetchResultListWithErrors(pairs, ErrEmptyOriginResponse)
	}
	if res.Error != nil {
		return fetchResultListWithErrors(pairs, res.Error)
	}

	// parse JSON
	var resp uniswapResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		return fetchResultListWithErrors(pairs, fmt.Errorf("failed to parse Uniswap response: %w", err))
	}

	// convert response from a slice to a map
	respMap := map[string]uniswapPairResponse{}
	for _, pairResp := range resp.Data.Pairs {
		respMap[pairResp.Token0.Symbol+"/"+pairResp.Token1.Symbol] = pairResp
	}

	// prepare result
	results := make([]FetchResult, 0)
	for _, pair := range pairs {
		b := pair.Base
		q := pair.Quote

		pair0 := b + "/" + q
		pair1 := q + "/" + b

		if r, ok := respMap[pair0]; ok {
			results = append(results, FetchResult{
				Price: Price{
					Pair:      pair,
					Price:     r.Price1.val(),
					Bid:       r.Price1.val(),
					Ask:       r.Price1.val(),
					Volume24h: r.Volume0.val(),
					Timestamp: time.Now(),
				},
			})
		} else if r, ok := respMap[pair1]; ok {
			results = append(results, FetchResult{
				Price: Price{
					Pair:      pair,
					Price:     r.Price0.val(),
					Bid:       r.Price0.val(),
					Ask:       r.Price0.val(),
					Volume24h: r.Volume1.val(),
					Timestamp: time.Now(),
				},
			})
		} else {
			results = append(results, FetchResult{
				Price: Price{Pair: pair},
				Error: ErrMissingResponseForPair,
			})
		}
	}

	return results
}
