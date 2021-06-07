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

package main

import (
	"crypto/ecdsa"
	"encoding/json"
	"log"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"

	"github.com/makerdao/oracle-suite/cmd/keeman/internal"
)

type eth struct {
	PrivateKey         *ecdsa.PrivateKey
	Password           string
	Account            accounts.Account
	Keystore           []byte
	showPass, showPriv bool
}

func (e eth) MarshalJSON() ([]byte, error) {
	type ks interface{}
	var k ks
	if err := json.Unmarshal(e.Keystore, &k); err != nil {
		return nil, err
	}

	var priv string
	if e.showPriv {
		priv = hexutil.Encode(crypto.FromECDSA(e.PrivateKey))
	}

	if e.showPass {
		return json.Marshal(struct {
			PrivateKey string           `json:"priv,omitempty"`
			Password   string           `json:"pass"`
			Account    accounts.Account `json:"account"`
			Keystore   ks               `json:"keystore"`
		}{
			PrivateKey: priv,
			Password:   e.Password,
			Account:    e.Account,
			Keystore:   k,
		})
	}

	return json.Marshal(struct {
		PrivateKey string           `json:"priv,omitempty"`
		Password   string           `json:"pass,omitempty"`
		Account    accounts.Account `json:"account"`
		Keystore   ks               `json:"keystore"`
	}{
		PrivateKey: priv,
		Account:    e.Account,
		Keystore:   k,
	})
}

func genEth(wallet *hdwallet.Wallet, path accounts.DerivationPath, pass string, showPass, showPriv bool) (*eth, error) {
	path = setPurpose(path, PathPurposeEth)
	log.Printf("eth.account path: %s", path)
	account, err := wallet.Derive(path, false)
	if err != nil {
		return nil, err
	}

	privateKey, err := wallet.PrivateKey(account)
	if err != nil {
		return nil, err
	}

	k, err := keystore.EncryptKey(internal.NewKey(privateKey), pass, keystore.StandardScryptN, keystore.StandardScryptP)
	if err != nil {
		return nil, err
	}

	e := &eth{
		Password:   pass,
		PrivateKey: privateKey,
		Account:    account,
		Keystore:   k,
		showPriv:   showPriv,
		showPass:   showPass,
	}
	return e, nil
}
