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
		PersistentPreRunE: func(_ *cobra.Command, args []string) error {
			var err error

			logger, err = newLogger(opts.Verbosity)
			if err != nil {
				return err
			}
			client, err = newClient(opts, logger)
			if err != nil {
				return err
			}

			return client.Start()
		},
		PersistentPostRunE: func(_ *cobra.Command, args []string) error {
			err := client.Stop()
			if err != nil {
				logger.WithError(err).Error("Unable to stop RPC Client")
			}
			return nil
		},
	}

	cmd.AddCommand(
		NewPullPricesCmd(),
		NewPullPriceCmd(),
	)

	return cmd
}

func NewPullPriceCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "price",
		Args:  cobra.ExactArgs(2),
		Short: "",
		Long:  ``,
		RunE: func(_ *cobra.Command, args []string) error {
			p, err := client.PullPrice(args[0], args[1])
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

func NewPullPricesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prices [PAIR,...] [ADDR,...]",
		Args:  cobra.MaximumNArgs(2),
		Short: "",
		Long:  ``,
		RunE: func(_ *cobra.Command, args []string) error {
			var filterPair, filterFrom string
			if len(args) > 0 {
				filterPair = args[0]
			}
			if len(args) > 1 {
				filterFrom = args[1]
			}

			p, err := client.PullPrices(filterPair, filterFrom)
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

	return cmd
}
