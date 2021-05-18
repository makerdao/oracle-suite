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
	"encoding/json"
	"flag"
	"fmt"

	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
)

func cmdGen(args []string) error {
	n, err := genFlags(args)
	if err != nil {
		return err
	}

	mnemonic, err := hdwallet.NewMnemonic(n * 8)
	if err != nil {
		return err
	}

	marshal, err := json.Marshal(mnemonic)
	if err != nil {
		return err
	}

	fmt.Println(string(marshal))

	return nil
}

func genFlags(args []string) (n int, err error) {
	fs := flag.NewFlagSet(args[0], flag.ExitOnError)

	fs.IntVar(&n, "bytes", 32, "Number of random bytes")

	err = fs.Parse(args[1:])

	return n, err
}
