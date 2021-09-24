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
	"math/big"
	"regexp"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/params"

	"github.com/makerdao/oracle-suite/internal/query"
	"github.com/makerdao/oracle-suite/pkg/ethereum"
)

const _CurvePoolJSON = `[{
"name":"get_dy",
"outputs":[{"type":"uint256","name":""}],
"inputs":[{"type":"int128","name":"i"},{"type":"int128","name":"j"},{"type":"uint256","name":"dx"}],
"stateMutability":"view",
"type":"function",
"gas":2654541
}]`

type CurveFinance struct {
	EthRPCURL                 string
	WorkerPool                query.WorkerPool
	ContractAddresses         ContractAddresses
	abi                       abi.ABI
	baseIndex, quoteIndex, dx *big.Int
}

func NewCurveFinance(
	ethRPCURL string,
	workerPool query.WorkerPool,
	contractAddresses ContractAddresses,
) (*CurveFinance, error) {

	a, err := abi.JSON(strings.NewReader(_CurvePoolJSON))
	if err != nil {
		return nil, err
	}
	return &CurveFinance{
		EthRPCURL:         ethRPCURL,
		WorkerPool:        workerPool,
		ContractAddresses: contractAddresses,
		abi:               a,
		baseIndex:         big.NewInt(0),
		quoteIndex:        big.NewInt(1),
		dx:                new(big.Int).Mul(big.NewInt(1), big.NewInt(params.Ether)),
	}, nil
}

func (s CurveFinance) pairsToContractAddress(pair Pair) (ethereum.Address, bool, error) {
	contract, inverted, ok := s.ContractAddresses.ByPair(pair)
	if !ok {
		return ethereum.Address{}, inverted, fmt.Errorf("failed to get contract address for pair: %s", pair.String())
	}
	return ethereum.HexToAddress(contract), inverted, nil
}

func (s CurveFinance) Pool() query.WorkerPool {
	return s.WorkerPool
}

func (s CurveFinance) PullPrices(pairs []Pair) []FetchResult {
	return callSinglePairOrigin(&s, pairs)
}

func (s CurveFinance) callOne(pair Pair) (*Price, error) {
	contract, inverted, err := s.pairsToContractAddress(pair)
	if err != nil {
		return nil, err
	}

	var callData []byte
	if !inverted {
		callData, err = s.abi.Pack("get_dy", s.baseIndex, s.quoteIndex, s.dx)
	} else {
		callData, err = s.abi.Pack("get_dy", s.quoteIndex, s.baseIndex, s.dx)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to pack contract args for pair: %s", pair.String())
	}

	price, err := ethCall(s.Pool(), s.EthRPCURL, contract, callData)
	if err != nil {
		return nil, err
	}

	return &Price{
		Pair:      pair,
		Price:     price,
		Timestamp: time.Now(),
	}, nil
}

func ethCall(
	workerPool query.WorkerPool,
	ethRPCURL string,
	contract ethereum.Address,
	callData []byte,
) (float64, error) {

	sprintf := fmt.Sprintf(
		`{"jsonrpc":"2.0","method":"eth_call","params":[{"to":"%s","data":"%s"},"latest"],"id":1}`,
		contract.String(),
		hexutil.Encode(callData),
	)
	res := workerPool.Query(&query.HTTPRequest{
		URL:    ethRPCURL,
		Method: "POST",
		Body:   bytes.NewBuffer([]byte(sprintf)),
	})
	if res.Error != nil {
		return 0, res.Error
	}
	if res == nil {
		return 0, ErrEmptyOriginResponse
	}

	var response jsonrpcMessage
	if err := json.Unmarshal(res.Body, &response); err != nil {
		return 0, fmt.Errorf("failed to parse response: %w", err)
	}
	if !response.isResponse() {
		return 0, ErrInvalidResponse
	}
	if response.Error != nil {
		return 0, fmt.Errorf("response error: %w", response.Error)
	}

	var result hexutil.Big
	if err := result.UnmarshalJSON(regexp.MustCompile(`0x[0]+`).ReplaceAll(response.Result, []byte("0x"))); err != nil {
		return 0, fmt.Errorf("failed to decode result: %w", err)
	}

	price, _ := new(big.Float).Quo(new(big.Float).SetInt(result.ToInt()), big.NewFloat(params.Ether)).Float64()
	return price, nil
}
