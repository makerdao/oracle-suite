//  Copyright (C) 2021 Maker Ecosystem Growth Holdings, INC.
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
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/makerdao/oracle-suite/cmd/keeman/internal"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"

	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
)

func der(q, mnemonic, pass, path string) error {
	seed := bip39.NewSeed(mnemonic, pass)
	masterKey, err := bip32.NewMasterKey(seed)
	if err != nil {
		return err
	}
	if q == "xprv" {
		fmt.Println(masterKey)
		return nil
	}

	publicKey := masterKey.PublicKey()
	if q == "xpub" {
		fmt.Println(publicKey)
		return nil
	}

	wallet, err := hdwallet.NewFromSeed(seed)
	if err != nil {
		return err
	}

	parsed, err := accounts.ParseDerivationPath(path)
	if err != nil {
		return err
	}

	account, err := wallet.Derive(parsed, false)
	if err != nil {
		return err
	}

	if q == "addr" {
		fmt.Println(account.Address.Hex())
		return nil
	}

	key, err := wallet.PrivateKey(account)
	if err != nil {
		return err
	}

	keyJson, err := keystore.EncryptKey(internal.NewKeyFromECDSA(key), pass, keystore.StandardScryptN, keystore.StandardScryptP)
	if err != nil {
		return err
	}
	if q == "json" {
		fmt.Println(string(keyJson))
		return nil
	}

	fmt.Println(" Master public key:", publicKey)
	fmt.Println("   Derivation path:", path)
	fmt.Println("  Ethereum address:", account.Address.Hex(), "https://etherscan.io/address/"+account.Address.Hex())
	fmt.Println("                   ", "https://etherscan.io/address/"+account.Address.Hex())

	return nil
}

func cmdDer(args []string) error {
	if internal.FileIsEmpty(os.Stdin) {
		return fmt.Errorf("missing mnemonic phrase")
	}

	fs := flag.NewFlagSet(args[0], flag.ExitOnError)
	var pass, path string
	fs.StringVar(&pass, "pass", "", "Seed password")
	fs.StringVar(&path, "path", "m/44'/60'/0'/0/0", "Derivation path")

	if err := fs.Parse(args[1:]); err != nil {
		return err
	}

	b, err := io.ReadAll(os.Stdin)
	if err != nil {
		return err
	}

	return der(fs.Arg(0), string(b), pass, path)
}
