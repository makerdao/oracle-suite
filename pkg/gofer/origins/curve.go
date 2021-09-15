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

	"github.com/ethereum/go-ethereum/accounts/abi"

	"github.com/makerdao/oracle-suite/internal/query"
)

type curveResponse struct {
	Data struct {
		Pairs []curvePairResponse
	}
}

type curveTokenResponse struct {
	Symbol string `json:"symbol"`
}

type curvePairResponse struct {
	ID      string             `json:"id"`
	Price0  stringAsFloat64    `json:"token0Price"`
	Price1  stringAsFloat64    `json:"token1Price"`
	Volume0 stringAsFloat64    `json:"volumeToken0"`
	Volume1 stringAsFloat64    `json:"volumeToken1"`
	Token0  curveTokenResponse `json:"token0"`
	Token1  curveTokenResponse `json:"token1"`
}

const _CurvePoolABI = `{"name":"get_dy","outputs":[{"type":"uint256","name":""}],"inputs":[{"type":"int128","name":"i"},{"type":"int128","name":"j"},{"type":"uint256","name":"dx"}],"stateMutability":"view","type":"function","gas":2654541}`

type CurveFinance struct {
	EthRpcUrl         string
	WorkerPool        query.WorkerPool
	ContractAddresses ContractAddresses
}

func NewCurveFinance(ethRpcUrl string, workerPool query.WorkerPool, contractAddresses ContractAddresses) CurveFinance {
	return CurveFinance{
		EthRpcUrl:         ethRpcUrl,
		WorkerPool:        workerPool,
		ContractAddresses: contractAddresses,
	}
}

func (s *CurveFinance) pairsToContractAddress(pair Pair) (string, error) {
	contract, ok := s.ContractAddresses.ByPair(pair)
	if !ok {
		return "", fmt.Errorf("failed to get Curve contract address for pair: %s", pair.String())
	}
	return contract, nil
}

func (s CurveFinance) Pool() query.WorkerPool {
	return s.WorkerPool
}

func (s CurveFinance) PullPrices(pairs []Pair) []FetchResult {
	return callSinglePairOrigin(&s, pairs)
}

// xnolint:dupl
func (s *CurveFinance) callOne(pair Pair) (*Price, error) {
	var err error

	contract, err := s.pairsToContractAddress(pair)
	if err != nil {
		return nil, err
	}

	abi.JSON(strings.NewReader(_CurvePoolABI))
	body := fmt.Sprintf(
		`{"jsonrpc":"2.0","method":"eth_call","params":[{"to":"%s","data":"%s"},"latest"],"id":1}`,
		contract,
	)

	req := &query.HTTPRequest{
		URL:    s.EthRpcUrl,
		Method: "POST",
		Body:   bytes.NewBuffer([]byte(body)),
	}

	// make query
	res := s.Pool().Query(req)
	if res == nil {
		return nil, ErrEmptyOriginResponse
	}
	if res.Error != nil {
		return nil, res.Error
	}

	// parse JSON
	var resp curveResponse
	err = json.Unmarshal(res.Body, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Curve response: %w", err)
	}

	// convert response from a slice to a map
	respMap := map[string]curvePairResponse{}
	for _, pairResp := range resp.Data.Pairs {
		respMap[pairResp.Token0.Symbol+"/"+pairResp.Token1.Symbol] = pairResp
	}

	b := pair.Base
	q := pair.Quote

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
