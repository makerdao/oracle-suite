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
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"

	"github.com/makerdao/oracle-suite/internal/query"
	"github.com/makerdao/oracle-suite/pkg/ethereum"
)

const _WrappedStakedETHJSON = `[{
"inputs":[],"name":"stEthPerToken","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],
"stateMutability":"view","type":"function"
},{
"inputs":[],"name":"tokensPerStEth","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],
"stateMutability":"view","type":"function"
},{
"inputs":[],"name":"totalSupply","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],
"stateMutability":"view","type":"function"
}]`

type WrappedStakedETH struct {
	EthRPCURL         string
	WorkerPool        query.WorkerPool
	ContractAddresses ContractAddresses
	abi               abi.ABI
}

func NewWrappedStakedETH(
	ethRPCURL string,
	workerPool query.WorkerPool,
	contractAddresses ContractAddresses,
) (*WrappedStakedETH, error) {

	a, err := abi.JSON(strings.NewReader(_WrappedStakedETHJSON))
	if err != nil {
		return nil, err
	}
	return &WrappedStakedETH{
		EthRPCURL:         ethRPCURL,
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

	var callData []byte
	if !inverted {
		callData, err = s.abi.Pack("stEthPerToken")
	} else {
		callData, err = s.abi.Pack("tokensPerStEth")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get contract args for pair: %s", pair.String())
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
