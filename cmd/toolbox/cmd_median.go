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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/spf13/cobra"

	"github.com/makerdao/gofer/pkg/ethereum"
	ethereumGeth "github.com/makerdao/gofer/pkg/ethereum/geth"
	"github.com/makerdao/gofer/pkg/oracle"
	oracleGeth "github.com/makerdao/gofer/pkg/oracle/geth"
	"github.com/makerdao/gofer/pkg/transport/messages"
)

type medianOptions struct {
	Address string
	Wat     string
}

var median oracle.Median

func NewMedianCmd(opts *options) *cobra.Command {
	var medianOpts medianOptions

	cmd := &cobra.Command{
		Use:   "median",
		Args:  cobra.ExactArgs(0),
		Short: "Commands related to the Medianizer contract",
		Long:  ``,
		PersistentPreRunE: func(_ *cobra.Command, args []string) error {
			var err error

			// Create signer:
			account, err := ethereumGeth.NewAccount(opts.EthereumKeystore, opts.EthereumPassword, ethereum.HexToAddress(opts.EthereumAddress))
			if err != nil {
				return err
			}
			signer := ethereumGeth.NewSigner(account)

			// Create Ethereum client:
			client, err := ethclient.Dial(opts.EthereumRPC)
			if err != nil {
				return err
			}
			gethClient := ethereumGeth.NewClient(client, signer)

			// Median instance:
			median = oracleGeth.NewMedian(gethClient, ethereum.HexToAddress(medianOpts.Address), medianOpts.Wat)

			return nil
		},
	}

	cmd.PersistentFlags().StringVar(
		&medianOpts.Address,
		"median-address",
		"",
		"Median contract address",
	)

	cmd.PersistentFlags().StringVar(
		&medianOpts.Wat,
		"median-wat",
		"",
		"Median contract asset name",
	)

	cmd.AddCommand(
		NewMedianAgeCmd(opts, &medianOpts),
		NewMedianBarCmd(opts, &medianOpts),
		NewMedianPriceCmd(opts, &medianOpts),
		NewMedianFeedsCmd(opts, &medianOpts),
		NewMedianPokeCmd(opts, &medianOpts),
		NewMedianLiftCmd(opts, &medianOpts),
		NewMedianDropCmd(opts, &medianOpts),
		NewMedianSetBarCmd(opts, &medianOpts),
	)

	return cmd
}

func NewMedianAgeCmd(opts *options, medianOpts *medianOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "age",
		Args:  cobra.ExactArgs(0),
		Short: "Returns contract age (last update time)",
		Long:  ``,
		RunE: func(_ *cobra.Command, args []string) error {
			age, err := median.Age(context.Background())
			if err != nil {
				return err
			}

			// Print age:
			fmt.Println(age.String())

			return nil
		},
	}
}

func NewMedianBarCmd(opts *options, medianOpts *medianOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "bar",
		Args:  cobra.ExactArgs(0),
		Short: "Returns contract bar (required quorum)",
		Long:  ``,
		RunE: func(_ *cobra.Command, args []string) error {
			bar, err := median.Bar(context.Background())
			if err != nil {
				return err
			}

			// Print bar:
			fmt.Println(bar)

			return nil
		},
	}
}

func NewMedianPriceCmd(opts *options, medianOpts *medianOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "val",
		Args:  cobra.ExactArgs(0),
		Short: "Returns oracle val (asset price)",
		Long:  ``,
		RunE: func(_ *cobra.Command, args []string) error {
			price, err := median.Val(context.Background())
			if err != nil {
				return err
			}

			// Print price:
			fmt.Println(price.String())

			return nil
		},
	}
}

func NewMedianFeedsCmd(opts *options, medianOpts *medianOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "feeds",
		Args:  cobra.ExactArgs(0),
		Short: "Returns list of feeds which are allowed to send prices",
		Long:  ``,
		RunE: func(_ *cobra.Command, args []string) error {
			feeds, err := median.Feeds(context.Background())
			if err != nil {
				return err
			}

			// Print feeds:
			for _, f := range feeds {
				fmt.Println(f.String())
			}

			return nil
		},
	}
}

func NewMedianPokeCmd(opts *options, medianOpts *medianOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "poke [json_messages_list]",
		Args:  cobra.ExactArgs(1),
		Short: "Directly invokes poke method",
		Long:  ``,
		RunE: func(_ *cobra.Command, args []string) error {
			var err error

			// Read JSON and parse it:
			in, err := readInput(args, 0)
			if err != nil {
				return err
			}

			msgs := &[]messages.Price{}
			err = json.Unmarshal(in, msgs)
			if err != nil {
				return err
			}

			var prices []*oracle.Price
			for _, m := range *msgs {
				prices = append(prices, m.Price)
			}

			tx, err := median.Poke(context.Background(), prices, true)
			if err != nil {
				return err
			}

			fmt.Printf("Transaction: %s\n", tx.String())

			return nil
		},
	}
}

func NewMedianLiftCmd(opts *options, medianOpts *medianOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "lift [addresses...]",
		Args:  cobra.MinimumNArgs(1),
		Short: "Adds given addresses to the feeders list",
		Long:  ``,
		RunE: func(_ *cobra.Command, args []string) error {
			var err error

			var addresses []ethereum.Address
			for _, a := range args {
				addresses = append(addresses, ethereum.HexToAddress(a))
			}

			tx, err := median.Lift(context.Background(), addresses, true)
			if err != nil {
				return err
			}

			fmt.Printf("Transaction: %s\n", tx.String())

			return nil
		},
	}
}

func NewMedianDropCmd(opts *options, medianOpts *medianOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "drop [addresses...]",
		Args:  cobra.MinimumNArgs(1),
		Short: "Removes given addresses from the feeders list",
		Long:  ``,
		RunE: func(_ *cobra.Command, args []string) error {
			var err error

			var addresses []ethereum.Address
			for _, a := range args {
				addresses = append(addresses, ethereum.HexToAddress(a))
			}

			tx, err := median.Drop(context.Background(), addresses, true)
			if err != nil {
				return err
			}

			fmt.Printf("Transaction: %s\n", tx.String())

			return nil
		},
	}
}

func NewMedianSetBarCmd(opts *options, medianOpts *medianOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "set-bar bar",
		Args:  cobra.ExactArgs(1),
		Short: "Sets bar variable (quorum)",
		Long:  ``,
		RunE: func(_ *cobra.Command, args []string) error {
			var err error

			bar, ok := (&big.Int{}).SetString(args[0], 10)
			if !ok {
				return errors.New("given value is not an valid number")
			}

			tx, err := median.SetBar(context.Background(), bar, true)
			if err != nil {
				return err
			}

			fmt.Printf("Transaction: %s\n", tx.String())

			return nil
		},
	}
}
