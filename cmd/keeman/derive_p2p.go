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
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/crypto"
	crypto2 "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
)

type p2p struct {
	Seed []byte
	ID   peer.ID
}

func (p p2p) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Seed string `json:"seed"`
		ID   string `json:"id"`
	}{Seed: hex.EncodeToString(p.Seed), ID: p.ID.String()})
}

func genP2p(wallet *hdwallet.Wallet, path accounts.DerivationPath) (*p2p, error) {
	path = setPurpose(path, PathPurposeP2p)
	log.Printf("p2p path: %s", path)
	privateKey, err := deriveKey(wallet, path)
	if err != nil {
		return nil, err
	}
	seed := crypto.FromECDSA(privateKey)[:32]

	privKey, err := peerPrivKey(seed)
	if err != nil {
		return nil, err
	}
	id, err := peer.IDFromPublicKey(privKey.GetPublic())
	if err != nil {
		return nil, err
	}

	return &p2p{Seed: seed, ID: id}, nil
}

func peerPrivKey(seed []byte) (crypto2.PrivKey, error) {
	if len(seed) != ed25519.SeedSize {
		return nil, fmt.Errorf("seed must be of size %d bytes - %d given", ed25519.SeedSize, len(seed))
	}
	privKey, _, err := crypto2.GenerateEd25519Key(bytes.NewReader(seed))
	if err != nil {
		return nil, err
	}
	return privKey, nil
}
