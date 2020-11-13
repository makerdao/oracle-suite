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
	"math/big"
)

type Transaction struct {
	// Address is the contract's address.
	Address Address
	// Nonce is the transaction nonce. If zero, the nonce will be filled
	// automatically.
	Nonce uint64
	// Gas is the gas price. If nil, the suggested gas price will be used.
	Gas *big.Int
	// GasLimit is the maximum gas available to be used for this transaction.
	GasLimit *big.Int
	// Data is the raw transaction data.
	Data []byte
	// ChainID is the transaction chain ID. If nil, the chan ID will be filled
	// automatically.
	ChainID *big.Int
	// SignedTx contains signed transaction. The data type stored here may
	// be different for various implementations.
	SignedTx interface{}
}

type Client interface {
	// Call executes a message call transaction, which is directly
	// executed in the VM of the node, but never mined into the blockchain.
	Call(ctx context.Context, address Address, data []byte) ([]byte, error)
	// Storage returns the value of key in the contract storage of the
	// given account.
	Storage(ctx context.Context, address Address, key Hash) ([]byte, error)
	// SendTransaction injects a signed transaction into the pending pool
	// for execution.
	SendTransaction(ctx context.Context, transaction *Transaction) (*Hash, error)
}
