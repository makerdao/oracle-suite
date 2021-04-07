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

const balancerURL = "https://api.thegraph.com/subgraphs/name/balancer-labs/balancer"

type balancerResponse struct {
	Data struct {
		TokenPrices []balancerPairResponse `json:"tokenPrices"`
	}
}

type balancerPairResponse struct {
	Symbol string          `json:"symbol"`
	Price  stringAsFloat64 `json:"price"`
	Volume stringAsFloat64 `json:"poolLiquidity"`
}

type Balancer struct {
	Pool query.WorkerPool
}

func (s *Balancer) pairsToContractAddress(pair Pair) string {
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

	p := Pair{Base: pair.Base, Quote: pair.Quote}

	switch {
	case match(p, Pair{Base: "BAL", Quote: "USD"}):
		return "0xba100000625a3754423978a60c9317c58a424e3d"
	case match(p, Pair{Base: "AAVE", Quote: "USD"}):
		return "0x7fc66500c84a76ad7e9c93437bfc5ac33e2ddae9"
	case match(p, Pair{Base: "WNXM", Quote: "USD"}):
		return "0x0d438f3b5175bebc262bf23753c1e53d03432bde"
	}

	return pair.String()
}

func (s *Balancer) Fetch(pairs []Pair) []FetchResult {
	return callSinglePairOrigin(s, pairs)
}

func (s *Balancer) callOne(pair Pair) (*Price, error) {
	var err error

	pairsJSON, _ := json.Marshal(s.pairsToContractAddress(pair))
	gql := `
		query($id:String) {
			tokenPrices(where: { id: $id }){
				symbol price poolLiquidity
			}
		}
	`
	body := fmt.Sprintf(
		`{"query":"%s","variables":{"id":%s}}`,
		strings.ReplaceAll(strings.ReplaceAll(gql, "\n", " "), "\t", ""),
		pairsJSON,
	)

	req := &query.HTTPRequest{
		URL:    balancerURL,
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
	var resp balancerResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Balancer response: %w", err)
	}

	pairPrice := resp.Data.TokenPrices[0]
	if pairPrice.Symbol != pair.Base {
		return nil, ErrMissingResponseForPair
	}

	return &Price{
		Pair:      pair,
		Price:     pairPrice.Price.val(),
		Volume24h: pairPrice.Volume.val(),
		Timestamp: time.Now(),
	}, nil
}
