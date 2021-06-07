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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/crypto"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
)

type caps struct {
	Shs  []byte
	Sign []byte
}

func (c caps) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Shs  string `json:"shs"`
		Sign string `json:"sign"`
	}{Shs: base64.URLEncoding.EncodeToString(c.Shs), Sign: base64.URLEncoding.EncodeToString(c.Sign)})
}

func genCaps(wallet *hdwallet.Wallet, path accounts.DerivationPath) (*caps, error) {
	iter, err := capsIterator(path)
	if err != nil {
		return nil, err
	}

	dpa := iter()
	log.Printf("caps.shs path: %s", dpa)
	a, err := deriveKey(wallet, dpa)
	if err != nil {
		return nil, err
	}

	dpb := iter()
	log.Printf("caps.sign path: %s", dpb)
	b, err := deriveKey(wallet, dpb)
	if err != nil {
		return nil, err
	}

	return &caps{Shs: crypto.FromECDSA(a), Sign: crypto.FromECDSA(b)}, nil
}

func capsIterator(base accounts.DerivationPath) (func() accounts.DerivationPath, error) {
	if len(base) < 3 {
		return nil, fmt.Errorf("derivation path needs at least three components")
	}

	path := make(accounts.DerivationPath, len(base)-1)
	copy(path[:], base[:])
	path[len(path)-2] = PathPurposeCaps
	path[len(path)-1] = 0

	return accounts.DefaultIterator(path), nil
}
