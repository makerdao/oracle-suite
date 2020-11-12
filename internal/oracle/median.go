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
	"math/big"
	"time"

	"github.com/makerdao/gofer/internal/ethereum"
)

// Median is an interface for the median oracle contract:
// https://github.com/makerdao/median/
//
// Contract documentation:
// https://docs.makerdao.com/smart-contract-modules/oracle-module/median-detailed-documentation
type Median interface {
	// Age returns the value from contract's age method. The age is the block
	// timestamp of last price val update.
	Age(ctx context.Context) (time.Time, error)
	// Bar returns the value from contract's bar method. The bar method returns
	// the minimum number of prices necessary to accept a new median value.
	Bar(ctx context.Context) (int64, error)
	// Price returns current asset price form the contract's storage.
	Price(ctx context.Context) (*big.Int, error)
	// Feeds returns a list of all Ethereum addresses that are authorized to update
	// Oracle prices.
	Feeds(ctx context.Context) ([]ethereum.Address, error)
	// Poke sends transaction to the smart contract which invokes contract's
	// poke method. If you set simulateBeforeRun to true, then transaction will be
	// simulated on the EVM before actual transaction will be send.
	Poke(ctx context.Context, prices []*Price, simulateBeforeRun bool) (*ethereum.Hash, error)
}
