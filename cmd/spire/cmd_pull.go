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

	"github.com/spf13/cobra"
)

func NewPullCmd(opts *options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pull",
		Args:  cobra.ExactArgs(1),
		Short: "",
		Long:  ``,
	}

	cmd.AddCommand(
		NewPullPriceCmd(opts),
		NewPullPricesCmd(opts),
	)

	return cmd
}

func NewPullPriceCmd(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:   "price",
		Args:  cobra.ExactArgs(2),
		Short: "",
		Long:  ``,
		RunE: func(_ *cobra.Command, args []string) error {
			ctx := context.Background()
			srv, err := PrepareClientServices(ctx, opts)
			if err != nil {
				return err
			}
			if err = srv.Start(); err != nil {
				return err
			}
			defer srv.CancelAndWait()

			p, err := srv.Client.PullPrice(args[0], args[1])
			if err != nil {
				return err
			}
			if p == nil {
				return errors.New("there is no price in the datastore for a given feeder and asset pair")
			}

			bts, err := json.Marshal(p)
			if err != nil {
				return err
			}

			fmt.Printf("%s\n", string(bts))

			return nil
		},
	}
}

type pullPricesOptions struct {
	FilterPair string
	FilterFrom string
}

func NewPullPricesCmd(opts *options) *cobra.Command {
	var pullPricesOpts pullPricesOptions

	cmd := &cobra.Command{
		Use:   "prices",
		Args:  cobra.ExactArgs(0),
		Short: "",
		Long:  ``,
		RunE: func(_ *cobra.Command, args []string) error {
			ctx := context.Background()
			srv, err := PrepareClientServices(ctx, opts)
			if err != nil {
				return err
			}
			if err = srv.Start(); err != nil {
				return err
			}
			defer srv.CancelAndWait()

			p, err := srv.Client.PullPrices(pullPricesOpts.FilterPair, pullPricesOpts.FilterFrom)
			if err != nil {
				return err
			}

			bts, err := json.Marshal(p)
			if err != nil {
				return err
			}

			fmt.Printf("%s\n", string(bts))

			return nil
		},
	}

	cmd.PersistentFlags().StringVar(
		&pullPricesOpts.FilterFrom,
		"filter.from",
		"",
		"",
	)

	cmd.PersistentFlags().StringVar(
		&pullPricesOpts.FilterPair,
		"filter.pair",
		"",
		"",
	)

	return cmd
}
