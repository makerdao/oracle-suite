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
	"crypto/ecdsa"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	crypto2 "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"

	"github.com/makerdao/oracle-suite/cmd/keeman/internal"
)

func der(mnemonic string, path accounts.DerivationPath, pass string) (*derOut, error) {
	wallet, err := hdwallet.NewFromMnemonic(mnemonic)
	if err != nil {
		return nil, err
	}

	iter, err := iteratorForGroup(path, 1)
	if err != nil {
		return nil, err
	}

	c, err := genCaps(wallet, iter)
	if err != nil {
		return nil, err
	}

	s, err := genSsb(wallet, chGrp(path, 3))
	if err != nil {
		return nil, err
	}

	p, err := genP2p(wallet, chGrp(path, 2))
	if err != nil {
		return nil, err
	}

	account, err := wallet.Derive(path, false)
	if err != nil {
		return nil, err
	}

	privateKey, err := wallet.PrivateKey(account)
	if err != nil {
		return nil, err
	}

	k, err := newKeystore(privateKey, pass)
	if err != nil {
		return nil, err
	}

	return &derOut{
		Path:     path,
		Prv:      hexutil.Encode(crypto.FromECDSA(privateKey)),
		Keystore: k,
		Caps:     c,
		P2p:      p,
		Ssb:      s,
	}, nil
}

type ks interface{}

func newKeystore(key *ecdsa.PrivateKey, pass string) (ks, error) {
	k, err := keystore.EncryptKey(internal.NewKey(key), pass, keystore.StandardScryptN, keystore.StandardScryptP)
	if err != nil {
		return nil, err
	}

	var x interface{}
	if err := json.Unmarshal(k, &x); err != nil {
		return nil, err
	}

	return x, nil
}

type derOut struct {
	Path     accounts.DerivationPath `json:"path"`
	Prv      string                  `json:"prv"`
	Keystore ks                      `json:"keystore"`
	Caps     *caps                   `json:"caps"`
	P2p      *p2p                    `json:"p2p"`
	Ssb      *ssb                    `json:"ssb"`
}

type p2p struct {
	Seed []byte  `json:"seed"`
	ID   peer.ID `json:"id"`
}

func (p p2p) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Seed string `json:"seed"`
		ID   string `json:"id"`
	}{Seed: hex.EncodeToString(p.Seed), ID: p.ID.String()})
}

func genP2p(wallet *hdwallet.Wallet, path accounts.DerivationPath) (*p2p, error) {
	privateKey, err := nextKey(wallet, path)
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

type caps struct {
	Shs  []byte `json:"shs"`
	Sign []byte `json:"sign"`
}

func (c caps) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Shs  string `json:"shs"`
		Sign string `json:"sign"`
	}{Shs: base64.URLEncoding.EncodeToString(c.Shs), Sign: base64.URLEncoding.EncodeToString(c.Sign)})
}

func genCaps(wallet *hdwallet.Wallet, iterate func() accounts.DerivationPath) (*caps, error) {
	a, err := nextKey(wallet, iterate())
	if err != nil {
		return nil, err
	}

	b, err := nextKey(wallet, iterate())
	if err != nil {
		return nil, err
	}
	return &caps{Shs: crypto.FromECDSA(a), Sign: crypto.FromECDSA(b)}, nil
}

type ssb struct {
	Type    string `json:"curve"`
	Public  []byte `json:"public"`
	Private []byte `json:"private"`
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
	k, err := nextKey(wallet, path)
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

func chGrp(base accounts.DerivationPath, group uint32) accounts.DerivationPath {
	path := make(accounts.DerivationPath, len(base))
	copy(path[:], base[:])
	path[len(path)-2] = group
	return path
}

func iteratorForGroup(base accounts.DerivationPath, group uint32) (func() accounts.DerivationPath, error) {
	if len(base) < 2 {
		return nil, fmt.Errorf("derivation path needs at least two components")
	}
	if base[len(base)-2] != 0 {
		return nil, fmt.Errorf("second to last path component needs to be 0")
	}

	path := make(accounts.DerivationPath, len(base))
	copy(path[:], base[:])
	path[len(path)-2] = group
	path[len(path)-1] = 0

	return accounts.DefaultIterator(path), nil
}

func nextKey(wallet *hdwallet.Wallet, path accounts.DerivationPath) (*ecdsa.PrivateKey, error) {
	account, err := wallet.Derive(path, false)
	if err != nil {
		return nil, err
	}
	return wallet.PrivateKey(account)
}

func cmdDer(args []string) error {
	mnemonic, path, pass, err := derInput(args, os.Stdin)
	if err != nil {
		return err
	}

	x, err := der(mnemonic, path, pass)
	if err != nil {
		return err
	}

	marshal, err := json.Marshal(x)
	if err != nil {
		return err
	}

	fmt.Println(string(marshal))

	return nil
}

func derInput(args []string, file *os.File) (string, accounts.DerivationPath, string, error) {
	fs := flag.NewFlagSet(args[0], flag.ExitOnError)

	var path, pass string
	fs.StringVar(&path, "path", "m/44'/60'/0'/0/0", "Derivation path")
	fs.StringVar(&pass, "pass", "", "Raw password or path to a file containing one")

	if err := fs.Parse(args[1:]); err != nil {
		return "", nil, "", err
	}

	parsedPath, err := accounts.ParseDerivationPath(path)
	if err != nil {
		return "", nil, "", err
	}

	if fileIsEmpty(file) {
		return "", nil, "", fmt.Errorf("missing mnemonic phrase")
	}

	mnemonic, err := io.ReadAll(file)
	if err != nil {
		return "", nil, "", err
	}

	return strings.Trim(string(mnemonic), "\t \n"), parsedPath, internal.ReadLineOrPass(pass), err
}
