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
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type Signer struct {
	wallet *Wallet
}

// NewSigner returns a new Signer instance. If you don't want to sign any data
// and only want to recover public addresses, you may use nil as an argument.
func NewSigner(wallet *Wallet) *Signer {
	return &Signer{
		wallet: wallet,
	}
}

func (s *Signer) Signature(data []byte) ([]byte, error) {
	msg := []byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(data), data))
	wallet := s.wallet.EthWallet()
	account := s.wallet.EthAccount()

	signature, err := wallet.SignDataWithPassphrase(*account, s.wallet.Passphrase(), "", msg)
	if err != nil {
		return nil, err
	}

	// Transform V from 0/1 to 27/28 according to the yellow paper:
	signature[64] += 27

	return signature, nil
}

func (s *Signer) Recover(signature []byte, data []byte) (*common.Address, error) {
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
