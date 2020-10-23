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
	"os"
	"runtime"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/signer/core"
)

type Wallet struct {
	accountManager *accounts.Manager
	passphrase     string
	address        common.Address
	wallet         accounts.Wallet
	account        *accounts.Account
}

func NewWallet(keyStore, passphrase string, address common.Address) (*Wallet, error) {
	var err error

	if keyStore == "" {
		keyStore = defaultKeyStore()
	}

	w := &Wallet{
		// Using StartClefAccountManager is not a perfect solution but it's probably little better than
		// copy-pasting the code.
		accountManager: core.StartClefAccountManager(keyStore, true, true, ""),
		passphrase:     passphrase,
		address:        address,
	}

	if w.wallet, w.account, err = w.findWalletByAddress(address); err != nil {
		return nil, err
	}

	return w, nil
}

func (s *Wallet) Address() common.Address {
	return s.address
}

func (s *Wallet) EthWallet() accounts.Wallet {
	return s.wallet
}

func (s *Wallet) EthAccount() *accounts.Account {
	return s.account
}

func (s *Wallet) Passphrase() string {
	return s.passphrase
}

func (s *Wallet) findWalletByAddress(from common.Address) (accounts.Wallet, *accounts.Account, error) {
	for _, wallet := range s.accountManager.Wallets() {
		for _, account := range wallet.Accounts() {
			if account.Address == from {
				return wallet, &account, nil
			}
		}
	}

	return nil, nil, errors.New("unable to find wallet for requested address")
}

// source: https://github.com/dapphub/dapptools/blob/master/src/ethsign/ethsign.go
func defaultKeyStore() string {
	var defaultKeyStores []string

	if runtime.GOOS == "darwin" {
		defaultKeyStores = []string{
			os.Getenv("HOME") + "/Library/Ethereum/keystore",
			os.Getenv("HOME") + "/Library/Application Support/io.parity.ethereum/keys/ethereum",
		}
	} else if runtime.GOOS == "windows" {
		// XXX: I'm not sure these paths are correct, but they are from geth/parity wikis.
		defaultKeyStores = []string{
			os.Getenv("APPDATA") + "/Ethereum/keystore",
			os.Getenv("APPDATA") + "/Parity/Ethereum/keys",
		}
	} else {
		defaultKeyStores = []string{
			os.Getenv("HOME") + "/.ethereum/keystore",
			os.Getenv("HOME") + "/.local/share/io.parity.ethereum/keys/ethereum",
		}
	}

	for _, keyStore := range defaultKeyStores {
		if _, err := os.Stat(keyStore); !os.IsNotExist(err) {
			return keyStore
		}
	}

	return ""
}
