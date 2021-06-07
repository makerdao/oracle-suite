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
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"log"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/crypto"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
)

type ssb struct {
	Type    string
	Public  []byte
	Private []byte
}

func (s ssb) MarshalJSON() ([]byte, error) {
	pub := base64.URLEncoding.EncodeToString(s.Public)
	return json.Marshal(struct {
		Curve   string `json:"curve"`
		Public  string `json:"public"`
		Private string `json:"private"`
		ID      string `json:"id"`
	}{
		Curve:   s.Type,
		Public:  pub + "." + s.Type,
		Private: base64.URLEncoding.EncodeToString(s.Private) + "." + s.Type,
		ID:      "@" + pub + "." + s.Type,
	})
}

func genSsb(wallet *hdwallet.Wallet, path accounts.DerivationPath) (*ssb, error) {
	path = setPurpose(path, PathPurposeSsb)
	log.Printf("ssb path: %s", path)
	k, err := deriveKey(wallet, path)
	if err != nil {
		return nil, err
	}
	a := ed25519.NewKeyFromSeed(crypto.FromECDSA(k))
	return &ssb{
		Type:    "ed25519",
		Private: a,
		Public:  a.Public().(ed25519.PublicKey),
	}, nil
}
