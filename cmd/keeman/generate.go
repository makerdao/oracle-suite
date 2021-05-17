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

	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
)

func gen(q string, bits int, pass string) error {
	entropy, err := bip39.NewEntropy(bits)
	if err != nil {
		return err
	}
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return err
	}

	if q == "mnemonic" {
		fmt.Println(mnemonic)
		return nil
	}

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

	fmt.Println("          Mnemonic:", mnemonic)
	fmt.Println("Master private key:", masterKey)
	fmt.Println(" Master public key:", publicKey)

	return nil
}

func cmdGen(args []string) error {
	fs := flag.NewFlagSet(args[0], flag.ExitOnError)
	var bits int
	fs.IntVar(&bits, "bits", 256, "Number of bits of entropy")
	var pass string
	fs.StringVar(&pass, "pass", "", "Seed password")

	if err := fs.Parse(args[1:]); err != nil {
		return err
	}

	return gen(fs.Arg(0), bits, pass)
}
