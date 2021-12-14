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
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/accounts"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"

	"github.com/chronicleprotocol/oracle-suite/cmd/keeman/internal"
)

const PathPurposeEth = 0
const PathPurposeP2p = 1
const PathPurposeSsb = 2
const PathPurposeCaps = 3

func der(mnemonic string, path accounts.DerivationPath, pass string, showPass, showPriv bool) (*derOut, error) {
	wallet, err := hdwallet.NewFromMnemonic(mnemonic)
	if err != nil {
		return nil, err
	}

	e, err := genEth(wallet, path, pass, showPass, showPriv)
	if err != nil {
		return nil, err
	}

	p, err := genP2p(wallet, path)
	if err != nil {
		return nil, err
	}
	s, err := genSsb(wallet, path)
	if err != nil {
		return nil, err
	}
	c, err := genCaps(wallet, path)
	if err != nil {
		return nil, err
	}

	return &derOut{
		Eth:  e,
		Caps: c,
		P2p:  p,
		Ssb:  s,
	}, nil
}

type derOut struct {
	Eth  *eth  `json:"eth"`
	Caps *caps `json:"caps"`
	P2p  *p2p  `json:"p2p"`
	Ssb  *ssb  `json:"ssb"`
}

func setPurpose(base accounts.DerivationPath, purpose uint32) accounts.DerivationPath {
	path := make(accounts.DerivationPath, len(base))
	copy(path[:], base[:])
	path[len(path)-3] = purpose
	return path
}

func deriveKey(wallet *hdwallet.Wallet, path accounts.DerivationPath) (*ecdsa.PrivateKey, error) {
	account, err := wallet.Derive(path, false)
	if err != nil {
		return nil, err
	}
	return wallet.PrivateKey(account)
}

func cmdDer(args []string) error {
	mnemonic, path, pass, showPass, showPriv, err := derInput(args, os.Stdin)
	if err != nil {
		return err
	}

	x, err := der(mnemonic, path, pass, showPass, showPriv)
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

func derInput(args []string, file *os.File) (string, accounts.DerivationPath, string, bool, bool, error) {
	fs := flag.NewFlagSet(args[0], flag.ExitOnError)

	var verbose bool
	fs.BoolVar(&verbose, "verbose", false, "Disable logging")

	var env, role, node int
	fs.IntVar(&env, "env", 0, "Environment index")
	fs.IntVar(&role, "role", 0, "Role index")
	fs.IntVar(&node, "node", 0, "Node index")

	var pass string
	fs.StringVar(&pass, "pass", "", "Raw password or path to a file containing one")

	var showPass, showPriv bool
	fs.BoolVar(&showPass, "showPass", false, "Include ethereum keystore password")
	fs.BoolVar(&showPriv, "showPriv", false, "Include ethereum private key")

	if err := fs.Parse(args[1:]); err != nil {
		return "", nil, "", showPass, showPriv, err
	}

	parsedPath, err := accounts.ParseDerivationPath(fmt.Sprintf("m/%d'/0/%d/%d", env, role, node))
	if err != nil {
		return "", nil, "", showPass, showPriv, err
	}

	if fileIsEmpty(file) {
		return "", nil, "", showPass, showPriv, fmt.Errorf("missing mnemonic phrase")
	}

	mnemonic, err := io.ReadAll(file)
	if err != nil {
		return "", nil, "", showPass, showPriv, err
	}

	if !verbose {
		log.SetOutput(io.Discard)
	}
	return strings.Trim(string(mnemonic), "\t \n"), parsedPath, internal.ReadLineOrSame(pass), showPass, showPriv, err
}
