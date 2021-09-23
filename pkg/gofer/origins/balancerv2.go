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
)

// The three values that can be queried:
//
// - PAIR_PRICE: the price of the tokens in the Pool, expressed as the price of the second token in units of the
//   first token. For example, if token A is worth $2, and token B is worth $4, the pair price will be 2.0.
//   Note that the price is computed *including* the tokens decimals. This means that the pair price of a Pool with
//   DAI and USDC will be close to 1.0, despite DAI having 18 decimals and USDC 6.
//
// - BPT_PRICE: the price of the Pool share token (BPT), in units of the first token.
//   Note that the price is computed *including* the tokens decimals. This means that the BPT price of a Pool with
//   USDC in which BPT is worth $5 will be 5.0, despite the BPT having 18 decimals and USDC 6.
//
// - INVARIANT: the value of the Pool's invariant, which serves as a measure of its liquidity.
// enum Variable { PAIR_PRICE, BPT_PRICE, INVARIANT }

const _BalancerV2PoolJSON = `[{
"inputs":[{"internalType":"enum IPriceOracle.Variable","name":"variable","type":"uint8"}],
"name":"getLatest",
"outputs":[{"internalType":"uint256","name":"","type":"uint256"}],
"stateMutability":"view","type":"function"
}]`

type BalancerV2 struct {
	EthRPCURL         string
	WorkerPool        query.WorkerPool
	ContractAddresses ContractAddresses
	abi               abi.ABI
	variable          byte
}

func NewBalancerV2(
	ethRPCURL string,
	workerPool query.WorkerPool,
	contractAddresses ContractAddresses,
) (*BalancerV2, error) {

	a, err := abi.JSON(strings.NewReader(_BalancerV2PoolJSON))
	if err != nil {
		return nil, err
	}
	return &BalancerV2{
		EthRPCURL:         ethRPCURL,
		WorkerPool:        workerPool,
		ContractAddresses: contractAddresses,
		abi:               a,
		variable:          0, // PAIR_PRICE
	}, nil
}

func (s BalancerV2) Pool() query.WorkerPool {
	return s.WorkerPool
}

func (s BalancerV2) PullPrices(pairs []Pair) []FetchResult {
	return callSinglePairOrigin(&s, pairs)
}

func (s BalancerV2) callOne(pair Pair) (*Price, error) {
	contract, inverted, err := s.ContractAddresses.AddressByPair(pair)
	if err != nil {
		return nil, err
	}
	if inverted {
		return nil, fmt.Errorf("cannot use inverted pair to retrieve price: %s", pair.String())
	}

	var callData []byte
	callData, err = s.abi.Pack("getLatest", s.variable)
	if err != nil {
		return nil, fmt.Errorf("failed to pack contract args for pair %s: %w", pair.String(), err)
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
