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
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/makerdao/gofer/internal/ethereum"
)

type Signer struct {
	account *Account
}

// NewSigner returns a new Signer instance. If you don't want to sign any data
// and you only want to recover public addresses, you may use nil as an argument.
func NewSigner(account *Account) *Signer {
	return &Signer{
		account: account,
	}
}

// Account implements the ethereum.Signer interface.
func (s *Signer) Address() ethereum.Address {
	if s.account == nil {
		return ethereum.Address{}
	}

	return s.account.address
}

// SignTransaction implements the ethereum.Signer interface.
func (s *Signer) SignTransaction(transaction *ethereum.Transaction) error {
	tx := types.NewTransaction(
		transaction.Nonce,
		transaction.Address,
		nil,
		transaction.GasLimit.Uint64(),
		transaction.Gas,
		transaction.Data,
	)
	signedTx, err := s.account.wallet.SignTxWithPassphrase(
		*s.account.account,
		s.account.passphrase,
		tx,
		transaction.ChainID,
	)
	if err != nil {
		return err
	}
	transaction.SignedTx = signedTx
	return nil
}

// Signature implements the ethereum.Signer interface.
func (s *Signer) Signature(data []byte) ([]byte, error) {
	return Signature(s.account, data)
}

// Recover implements the ethereum.Signer interface.
func (s *Signer) Recover(signature []byte, data []byte) (*ethereum.Address, error) {
	return Recover(signature, data)
}

func Signature(account *Account, data []byte) ([]byte, error) {
	msg := []byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(data), data))

	signature, err := account.wallet.SignDataWithPassphrase(*account.account, account.passphrase, "", msg)
	if err != nil {
		return nil, err
	}

	// Transform V from 0/1 to 27/28 according to the yellow paper:
	signature[64] += 27

	return signature, nil
}

func Recover(signature []byte, data []byte) (*ethereum.Address, error) {
	if len(signature) != 65 {
		return nil, errors.New("signature must be 65 bytes long")
	}
	if signature[64] != 27 && signature[64] != 28 {
		return nil, errors.New("invalid Ethereum signature (V is not 27 or 28)")
	}

	// Transform yellow paper V from 27/28 to 0/1:
	signature[64] -= 27

	msg := []byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(data), data))
	hash := crypto.Keccak256(msg)

	rpk, err := crypto.SigToPub(hash, signature)
	if err != nil {
		return nil, err
	}

	address := crypto.PubkeyToAddress(*rpk)
	return &address, nil
}
