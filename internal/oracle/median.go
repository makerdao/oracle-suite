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

package oracle

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sort"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/makerdao/gofer/internal/ethereum"
)

// TODO: make it configurable
const gasLimit = 200000
const maxReadRetries = 3
const delayBetweenReadRetries = 5 * time.Second

// Median is an interface for the median oracle contract:
// https://github.com/makerdao/median/
//
// Contract documentation:
// https://docs.makerdao.com/smart-contract-modules/oracle-module/median-detailed-documentation
type Median struct {
	ethereum  *ethereum.Client
	address   common.Address
	assetPair string
}

// NewMedian creates the new Median instance.
func NewMedian(ethereum *ethereum.Client, address common.Address, assetPair string) *Median {
	return &Median{
		ethereum:  ethereum,
		address:   address,
		assetPair: assetPair,
	}
}

// Age returns the value from contract's age method. The age is the block
// timestamp of last price val update.
func (m *Median) Age(ctx context.Context) (time.Time, error) {
	r, err := m.read(ctx, "age")
	if err != nil {
		return time.Unix(0, 0), err
	}

	return time.Unix(int64(r[0].(uint32)), 0), nil
}

// Bar returns the value from contract's bar method. The bar method returns
// the minimum number of prices necessary to accept a new median value.
func (m *Median) Bar(ctx context.Context) (int64, error) {
	r, err := m.read(ctx, "bar")
	if err != nil {
		return 0, err
	}

	return r[0].(*big.Int).Int64(), nil
}

// Price returns current asset price form the contract's storage.
func (m *Median) Price(ctx context.Context) (*big.Int, error) {
	b, err := m.ethereum.Storage(ctx, m.address, common.BigToHash(big.NewInt(1)))
	if err != nil {
		return nil, err
	}
	if len(b) < 32 {
		return nil, errors.New("oracle contract storage query failed")
	}

	return new(big.Int).SetBytes(b[16:32]), err
}

// AuthorizedOracles returns a list of all authorized Oracle addresses.
// Note, that this method need to execute 256 calls to the ethereum client
// to fetch all addresses.
func (m *Median) AuthorizedOracles(ctx context.Context) ([]common.Address, error) {
	var orcl []common.Address
	var null common.Address

	mu := sync.Mutex{}
	wg := sync.WaitGroup{}

	wg.Add(256)
	for i := 0; i < 256; i++ {
		i := i
		go func() {
			defer wg.Done()
			addr, err := m.read(ctx, "slot", uint8(i))
			if err != nil {
				return
			}
			if len(addr) != 1 {
				return
			}
			if addr := addr[0].(common.Address); addr != null {
				mu.Lock()
				orcl = append(orcl, addr)
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	return orcl, nil
}

// Poke sends transaction to the smart contract which invokes contract's
// poke method. If you set simulateBeforeRun to true, then transaction will be
// simulated on the EVM before actual transaction will be send.
func (m *Median) Poke(ctx context.Context, prices []*Price, simulateBeforeRun bool) (*common.Hash, error) {
	// It's important to send prices in correct order, otherwise contract will fail:
	sort.Slice(prices, func(i, j int) bool {
		return prices[i].Val.Cmp(prices[j].Val) < 0
	})

	var (
		val []*big.Int
		age []*big.Int
		v   []uint8
		r   [][32]byte
		s   [][32]byte
	)

	for _, arg := range prices {
		if arg.AssetPair != m.assetPair {
			return nil, fmt.Errorf(
				"incompatible asset pair, %s given but %s expected",
				arg.AssetPair,
				m.assetPair,
			)
		}

		val = append(val, arg.Val)
		age = append(age, big.NewInt(arg.Age.Unix()))
		v = append(v, arg.V)
		r = append(r, arg.R)
		s = append(s, arg.S)
	}

	// Simulate transaction to not waste a gas in case of an error:
	if simulateBeforeRun {
		if _, err := m.read(ctx, "poke", val, age, v, r, s); err != nil {
			return nil, err
		}
	}

	return m.write(ctx, "poke", val, age, v, r, s)
}

func (m *Median) read(ctx context.Context, method string, args ...interface{}) ([]interface{}, error) {
	cd, err := medianABI.Pack(method, args...)
	if err != nil {
		return nil, err
	}

	var data []byte
	err = retry(maxReadRetries, delayBetweenReadRetries, func() error {
		data, err = m.ethereum.Call(ctx, m.address, cd)
		return err
	})
	if err != nil {
		return nil, err
	}

	return medianABI.Unpack(method, data)
}

func (m *Median) write(ctx context.Context, method string, args ...interface{}) (*common.Hash, error) {
	cd, err := medianABI.Pack(method, args...)
	if err != nil {
		return nil, err
	}

	return m.ethereum.SendTransaction(ctx, m.address, gasLimit, cd)
}

func retry(maxRetries int, delay time.Duration, f func() error) error {
	for i := 0; ; i++ {
		err := f()
		if err == nil {
			return err
		}
		if i >= (maxRetries - 1) {
			break
		}
		time.Sleep(delay)
	}
	return nil
}
