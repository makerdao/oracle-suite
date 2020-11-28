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

package mocks

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/mock"
)

type EthClient struct {
	mock.Mock
}

func (e *EthClient) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	args := e.Called(ctx, tx)
	return args.Error(0)
}

func (e *EthClient) StorageAt(ctx context.Context, acc common.Address, key common.Hash, block *big.Int) ([]byte, error) {
	args := e.Called(ctx, acc, key, block)
	return args.Get(0).([]byte), args.Error(1)
}

func (e *EthClient) CallContract(ctx context.Context, call ethereum.CallMsg, block *big.Int) ([]byte, error) {
	args := e.Called(ctx, call, block)
	return args.Get(0).([]byte), args.Error(1)
}

func (e *EthClient) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	args := e.Called(ctx, account)
	return uint64(args.Int(0)), args.Error(1)
}

func (e *EthClient) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	args := e.Called(ctx)
	return args.Get(0).(*big.Int), args.Error(1)
}

func (e *EthClient) NetworkID(ctx context.Context) (*big.Int, error) {
	args := e.Called(ctx)
	return args.Get(0).(*big.Int), args.Error(1)
}
