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

package ethereum

import (
	"context"
)

// Client is a wrapper for the Ethereum client.
type Client interface {
	// Call executes a message call transaction, which is directly
	// executed in the VM  of the node, but never mined into the blockchain.
	Call(ctx context.Context, address Address, data []byte) ([]byte, error)
	// Storage returns the value of key in the contract storage of the
	// given account.
	Storage(ctx context.Context, address Address, key Hash) ([]byte, error)
	// SendTransaction injects a signed transaction into the pending pool
	// for execution. It's uses suggested gas price by default.
	SendTransaction(ctx context.Context, address Address, gasLimit uint64, data []byte) (*Hash, error)
}
