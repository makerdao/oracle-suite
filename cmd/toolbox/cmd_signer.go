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
	"encoding/hex"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/makerdao/gofer/pkg/ethereum"
	"github.com/makerdao/gofer/pkg/ethereum/geth"
)

type signerOptions struct {
	Hex bool
}

var signer ethereum.Signer

func NewSignerCmd(opts *options) *cobra.Command {
	var signerOpts signerOptions

	cmd := &cobra.Command{
		Use:   "signer",
		Args:  cobra.ExactArgs(0),
		Short: "Commands used to sign and verify data",
		Long:  ``,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if opts.EthereumAddress != "" {
				account, err := geth.NewAccount(opts.EthereumKeystore, opts.EthereumPassword, ethereum.HexToAddress(opts.EthereumAddress))
				if err != nil {
					return err
				}

				signer = geth.NewSigner(account)
			} else {
				signer = geth.NewSigner(nil)
			}
			return nil
		},
	}

	cmd.PersistentFlags().BoolVar(
		&signerOpts.Hex,
		"hex",
		false,
		"Is input encoded as a string",
	)

	cmd.AddCommand(
		NewSignerSignCmd(opts, &signerOpts),
		NewSignerVerifyCmd(opts, &signerOpts),
	)

	return cmd
}

func NewSignerSignCmd(opts *options, signerOpts *signerOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "sign [input]",
		Args:  cobra.MaximumNArgs(1),
		Short: "Signs given input (stdin is used if input argument is empty)",
		Long:  ``,
		RunE: func(_ *cobra.Command, args []string) error {
			var err error

			in, err := readInput(args, 0)
			if signerOpts.Hex {
				in, err = hex.DecodeString(string(in))
				if err != nil {
					return err
				}
			}

			signature, err := signer.Signature(in)
			if err != nil {
				return err
			}

			fmt.Printf("%x\n", signature.Bytes())

			return nil
		},
	}
}

func NewSignerVerifyCmd(opts *options, signerOpts *signerOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "verify signature [input]",
		Args:  cobra.MinimumNArgs(1),
		Short: "Verifies given signature (stdin is used if input argument is empty)",
		Long:  ``,
		RunE: func(_ *cobra.Command, args []string) error {
			var err error

			in, err := readInput(args, 1)
			if signerOpts.Hex {
				in, err = hex.DecodeString(string(in))
				if err != nil {
					return err
				}
			}

			signature, err := hex.DecodeString(args[0])
			if err != nil {
				return err
			}

			address, err := signer.Recover(ethereum.SignatureFromBytes(signature), in)
			if err != nil {
				return err
			}

			fmt.Println(address.String())

			return nil
		},
	}
}
