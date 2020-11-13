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

package geth

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"regexp"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"

	internalEthereum "github.com/makerdao/gofer/internal/ethereum"
)

// RevertErr may be returned by Client.Call method in case of EVN revert.
type RevertErr struct {
	Message string
	Err     error
}

func (e RevertErr) Error() string {
	return fmt.Sprintf("reverted: %s", e.Message)
}

func (e RevertErr) Unwrap() error {
	return e.Err
}

// ethClient represents the Ethereum client, like the ethclient.Client.
type ethClient interface {
	ethereum.TransactionSender
	ethereum.ChainStateReader
	ethereum.ContractCaller
	ethereum.PendingStateReader
	ethereum.GasPricer

	NetworkID(ctx context.Context) (*big.Int, error)
}

// Client implements the ethereum.Client interface.
type Client struct {
	ethClient ethClient
	signer    internalEthereum.Signer
}

// NewClient returns a new Client instance.
func NewClient(ethClient ethClient, signer internalEthereum.Signer) *Client {
	return &Client{
		ethClient: ethClient,
		signer:    signer,
	}
}

// Call implements the ethereum.Client interface.
func (e *Client) Call(ctx context.Context, address internalEthereum.Address, data []byte) ([]byte, error) {
	cm := ethereum.CallMsg{
		From:     e.signer.Address(),
		To:       &address,
		Gas:      0,
		GasPrice: nil,
		Value:    nil,
		Data:     data,
	}

	resp, err := e.ethClient.CallContract(ctx, cm, nil)
	if err := isRevertErr(err); err != nil {
		return nil, err
	}
	if err := isRevertResp(resp); err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	return resp, err
}

// Storage implements the ethereum.Client interface.
func (e *Client) Storage(ctx context.Context, address internalEthereum.Address, key internalEthereum.Hash) ([]byte, error) {
	return e.ethClient.StorageAt(ctx, address, key, nil)
}

// SendTransaction implements the ethereum.Client interface.
func (e *Client) SendTransaction(ctx context.Context, transaction *internalEthereum.Transaction) (*internalEthereum.Hash, error) {
	var err error

	tx := &internalEthereum.Transaction{
		Address:  transaction.Address,
		Nonce:    transaction.Nonce,
		Gas:      new(big.Int).Set(transaction.Gas),
		GasLimit: new(big.Int).Set(transaction.GasLimit),
		SignedTx: transaction.SignedTx,
	}

	copy(tx.Data, transaction.Data)

	if tx.Nonce == 0 {
		tx.Nonce, err = e.ethClient.PendingNonceAt(ctx, e.signer.Address())
		if err != nil {
			return nil, err
		}
	}

	if tx.Gas == nil {
		tx.Gas, err = e.ethClient.SuggestGasPrice(ctx)
		if err != nil {
			return nil, err
		}
	}

	if tx.ChainID == nil {
		tx.ChainID, err = e.ethClient.NetworkID(ctx)
		if err != nil {
			return nil, err
		}
	}

	if tx.SignedTx == nil {
		err = e.signer.SignTransaction(tx)
		if err != nil {
			return nil, err
		}
	}

	if stx, ok := tx.SignedTx.(*types.Transaction); ok {
		hash := stx.Hash()
		return &hash, e.ethClient.SendTransaction(ctx, stx)
	}

	return nil, errors.New("unable to send transaction, SignedTx field have invalid type")
}

func isRevertResp(resp []byte) error {
	revert, err := abi.UnpackRevert(resp)
	if err != nil {
		return nil
	}

	return RevertErr{Message: revert, Err: nil}
}

func isRevertErr(vmErr error) error {
	switch terr := vmErr.(type) {
	case rpc.DataError:
		// Some RPC servers returns "revert" data as a hex encoded string, here
		// we're trying to parse it:
		if str, ok := terr.ErrorData().(string); ok {
			re := regexp.MustCompile("(0x[a-zA-Z0-9]+)")
			match := re.FindStringSubmatch(str)

			if len(match) == 2 && len(match[1]) > 2 {
				bytes, err := hex.DecodeString(match[1][2:])
				if err != nil {
					return nil
				}

				revert, err := abi.UnpackRevert(bytes)
				if err != nil {
					return nil
				}

				return RevertErr{Message: revert, Err: vmErr}
			}
		}
	}

	return nil
}
