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
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/makerdao/gofer/cmd/gofer/internal/marshal"
	"github.com/makerdao/gofer/pkg/cli"
	"github.com/makerdao/gofer/pkg/gofer"
)

func NewPairsCmd(o *options) *cobra.Command {
	return &cobra.Command{
		Use:   "pairs",
		Args:  cobra.NoArgs,
		Short: "List all supported pairs",
		Long:  `List all supported asset pairs.`,
		RunE: func(_ *cobra.Command, args []string) error {
			m, err := marshal.NewMarshal(o.OutputFormat.format)
			if err != nil {
				return err
			}

			return cli.Pairs(o.ConfigFilePath, m)
		},
	}
}

func NewExchangesCmd(o *options) *cobra.Command {
	return &cobra.Command{
		Use:   "exchanges [PAIR...]",
		Short: "List supported exchanges",
		Long: `Lists exchanges that will be queried for all of the supported pairs
or a subset of those, if at least one PAIR is provided.`,
		RunE: func(_ *cobra.Command, args []string) error {
			m, err := marshal.NewMarshal(o.OutputFormat.format)
			if err != nil {
				return err
			}

			absPath, err := filepath.Abs(o.ConfigFilePath)
			if err != nil {
				panic(err)
			}

			l, err := gofer.ReadFile(absPath)
			if err != nil {
				panic(err)
			}

			return cli.Exchanges(args, l, m)
		},
	}
}

func NewPriceCmd(o *options) *cobra.Command {
	return &cobra.Command{
		Use:   "price PAIR [PAIR...]",
		Args:  cobra.MinimumNArgs(1),
		Short: "Return price for given PAIRs",
		Long:  `Print the price of given PAIRs`,
		RunE: func(_ *cobra.Command, args []string) error {
			m, err := marshal.NewMarshal(o.OutputFormat.format)
			if err != nil {
				return err
			}

			absPath, err := filepath.Abs(o.ConfigFilePath)
			if err != nil {
				panic(err)
			}

			l, err := gofer.ReadFile(absPath)
			if err != nil {
				panic(err)
			}

			return cli.Price(args, l, m)
		},
	}
}

func NewRootCommand(opts *options) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "gofer",
		Version: "DEV",
		Short:   "Tool for providing reliable data in the blockchain ecosystem",
		Long: `
Gofer is a CLI interface for the Gofer Go Library.

It is a tool that allows for easy data retrieval from various sources
with aggregates that increase reliability in the DeFi environment.`,
	}

	rootCmd.PersistentFlags().StringVarP(&opts.ConfigFilePath, "config", "c", "./gofer.json", "config file")
	rootCmd.PersistentFlags().VarP(&opts.OutputFormat, "format", "f", "output format")

	return rootCmd
}
