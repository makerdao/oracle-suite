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

import "github.com/spf13/cobra"

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
		"ethereum keystore path",
	)

	rootCmd.PersistentFlags().StringVar(
		&opts.EthereumPassword,
		"eth-password",
		"",
		"ethereum keystore password",
	)

	rootCmd.PersistentFlags().StringVar(
		&opts.EthereumAddress,
		"eth-address",
		"",
		"ethereum account address",
	)

	rootCmd.PersistentFlags().StringVar(
		&opts.EthereumRPC,
		"eth-rpc",
		"",
		"ethereum RPC address",
	)

	rootCmd.AddCommand(
		NewMedianCmd(&opts),
		NewPriceCmd(&opts),
		NewSignerCmd(&opts),
	)

	return rootCmd
}
