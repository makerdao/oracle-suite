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
	"encoding/base64"
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
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"

	"github.com/makerdao/oracle-suite/cmd/keeman/internal"
)

func der(mnemonic string, path accounts.DerivationPath, pass string) (*derOut, error) {
	wallet, err := hdwallet.NewFromMnemonic(mnemonic)
	if err != nil {
		return nil, err
	}

	c, err := genCaps(wallet, path)
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

	k, err := genKeystore(privateKey, pass)
	if err != nil {
		return nil, err
	}

	var x interface{}
	if err := json.Unmarshal(k, &x); err != nil {
		return nil, err
	}
	return &derOut{
		path.String(),
		hexutil.Encode(crypto.FromECDSA(privateKey)),
		x,
		*c,
	}, nil
}

type derOut struct {
	Path     string      `json:"path"`
	Prv      string      `json:"prv"`
	Keystore interface{} `json:"keystore"`
	Caps     caps        `json:"caps"`
}
type caps struct {
	Shs  string `json:"shs"`
	Sign string `json:"sign"`
}

func genCaps(wallet *hdwallet.Wallet, base accounts.DerivationPath) (*caps, error) {
	if len(base) < 2 {
		return nil, fmt.Errorf("derivation path needs at least two components")
	}
	if base[len(base)-2] != 0 {
		return nil, fmt.Errorf("second to last path component needs to be 0")
	}

	path := make(accounts.DerivationPath, len(base))
	copy(path[:], base[:])
	path[len(path)-2] = 1
	path[len(path)-1] = 0
	f := accounts.DefaultIterator(path)

	a, err := nextKey(wallet, f())
	if err != nil {
		return nil, err
	}

	b, err := nextKey(wallet, f())
	if err != nil {
		return nil, err
	}

	return &caps{
		base64.URLEncoding.EncodeToString(crypto.FromECDSA(a)),
		base64.URLEncoding.EncodeToString(crypto.FromECDSA(b)),
	}, nil
}

func nextKey(wallet *hdwallet.Wallet, path accounts.DerivationPath) (*ecdsa.PrivateKey, error) {
	account, err := wallet.Derive(path, false)
	if err != nil {
		return nil, err
	}

	key, err := wallet.PrivateKey(account)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func genKeystore(key *ecdsa.PrivateKey, pass string) ([]byte, error) {
	return keystore.EncryptKey(internal.NewKey(key), pass, keystore.StandardScryptN, keystore.StandardScryptP)
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
