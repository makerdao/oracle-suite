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
	"log"
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

const _WrappedStakedETHJSON = `[
{"inputs":[],"name":"stEthPerToken","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},
{"inputs":[],"name":"tokensPerStEth","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},
{"inputs":[],"name":"totalSupply","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"}
]`

type WrappedStakedETH struct {
	EthRpcUrl         string
	WorkerPool        query.WorkerPool
	ContractAddresses ContractAddresses
	abi               abi.ABI
}

func NewWrappedStakedETH(ethRpcUrl string, workerPool query.WorkerPool, contractAddresses ContractAddresses) (*WrappedStakedETH, error) {
	a, err := abi.JSON(strings.NewReader(_WrappedStakedETHJSON))
	if err != nil {
		return nil, err
	}
	return &WrappedStakedETH{
		EthRpcUrl:         ethRpcUrl,
		WorkerPool:        workerPool,
		ContractAddresses: contractAddresses,
		abi:               a,
	}, nil
}

func (s WrappedStakedETH) pairsToContractAddress(pair Pair) (ethereum.Address, bool, error) {
	contract, inverted, ok := s.ContractAddresses.ByPair(pair)
	if !ok {
		return ethereum.Address{}, inverted, fmt.Errorf("failed to get Curve contract address for pair: %s", pair.String())
	}
	return ethereum.HexToAddress(contract), inverted, nil
}

func (s WrappedStakedETH) Pool() query.WorkerPool {
	return s.WorkerPool
}

func (s WrappedStakedETH) PullPrices(pairs []Pair) []FetchResult {
	return callSinglePairOrigin(&s, pairs)
}

func (s WrappedStakedETH) callOne(pair Pair) (*Price, error) {
	contract, inverted, err := s.pairsToContractAddress(pair)
	if err != nil {
		return nil, err
	}

	var args []byte
	if !inverted {
		args, err = s.abi.Pack("stEthPerToken")
	} else {
		args, err = s.abi.Pack("tokensPerStEth")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get Curve contract args for pair: %s", pair.String())
	}

	res := s.Pool().Query(&query.HTTPRequest{
		URL:    s.EthRpcUrl,
		Method: "POST",
		Body: bytes.NewBuffer([]byte(fmt.Sprintf(
			`{"jsonrpc":"2.0","method":"eth_call","params":[{"to":"%s","data":"%s"},"latest"],"id":1}`,
			contract.String(),
			hexutil.Encode(args),
		))),
	})
	if res.Error != nil {
		return nil, res.Error
	}
	if res == nil {
		return nil, ErrEmptyOriginResponse
	}

	var response jsonrpcMessage
	err = json.Unmarshal(res.Body, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Curve response: %w", err)
	}
	if response.Error != nil {
		return nil, response.Error
	}

	var result hexutil.Big
	if err := result.UnmarshalJSON(regexp.MustCompile(`0x[0]+`).ReplaceAll(response.Result, []byte("0x"))); err != nil {
		return nil, fmt.Errorf("failed to decode Curve result: %w", err)
	}

	price, _ := new(big.Float).Quo(new(big.Float).SetInt(result.ToInt()), big.NewFloat(params.Ether)).Float64()

	log.Println("wsteth", inverted, price)

	return &Price{
		Pair:      pair,
		Price:     price,
		Timestamp: time.Now(),
	}, nil
}
