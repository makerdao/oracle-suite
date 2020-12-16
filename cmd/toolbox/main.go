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
	"io"
	"os"
	"sort"

	"github.com/spf13/cobra"

	"github.com/makerdao/gofer/pkg/log"
)

func main() {
	rootCmd := NewRootCommand()

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

type options struct {
	EthereumKeystore string
	EthereumPassword string
	EthereumAddress  string
	EthereumRPC      string
}


func NewRootCommand() *cobra.Command {
	var opts options
	
	rootCmd := &cobra.Command{
		Use:           "toolbox",
		Version:       "DEV",
		Short:         "",
		Long:          ``,
		SilenceErrors: false,
		SilenceUsage:  true,
	}

	rootCmd.PersistentFlags().StringVar(
		&opts.EthereumKeystore,
		"eth-keystore",
		"",
		"Ethereum keystore path",
	)

	rootCmd.PersistentFlags().StringVar(
		&opts.EthereumPassword,
		"eth-password",
		"",
		"Ethereum keystore password",
	)

	rootCmd.PersistentFlags().StringVar(
		&opts.EthereumAddress,
		"eth-address",
		"",
		"Ethereum account address",
	)

	rootCmd.PersistentFlags().StringVar(
		&opts.EthereumRPC,
		"eth-rpc",
		"",
		"Ethereum RPC address",
	)

	rootCmd.AddCommand(
		NewMedianCmd(&opts),
		NewPriceCmd(&opts),
		NewSignerCmd(&opts),
	)

	return rootCmd
}

func readAll(r io.Reader) ([]byte, error) {
	b := make([]byte, 0, 512)
	for {
		if len(b) == cap(b) {
			// Add more capacity (let append pick how much).
			b = append(b, 0)[:len(b)]
		}
		n, err := r.Read(b[len(b):cap(b)])
		b = b[:len(b)+n]
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return b, err
		}
	}
}

func readInput(args []string, pos int) ([]byte, error) {
	var err error

	in := os.Stdin
	if len(args) > pos {
		in, err = os.Open(args[pos])
		if err != nil {
			return nil, err
		}
	}

	return readAll(in)
}

func sortFields(f log.Fields) []string {
	var ks []string
	for k := range f {
		ks = append(ks, k)
	}

	sort.Strings(ks)

	return ks
}
