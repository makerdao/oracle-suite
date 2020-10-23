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
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/makerdao/gofer/pkg/relayer"
	"github.com/makerdao/gofer/pkg/relayer/config"
)

func newRelayer(path string) (*relayer.Relayer, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	j, err := config.ParseJSONFile(absPath)
	if err != nil {
		return nil, err
	}

	r, err := j.MakeRelayer()
	if err != nil {
		return nil, err
	}
	return r, nil
}

func NewRunCmd(o *options) *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Args:  cobra.ExactArgs(0),
		Short: "",
		Long:  ``,
		RunE: func(_ *cobra.Command, _ []string) error {
			absPath, err := filepath.Abs(o.ConfigFilePath)
			if err != nil {
				return err
			}

			rel, err := newRelayer(absPath)
			if err != nil {
				return err
			}

			err = rel.Start(nil, nil)
			if err != nil {
				return err
			}
			defer rel.Stop()

			c := make(chan os.Signal)
			signal.Notify(c, os.Interrupt, syscall.SIGTERM)
			<-c

			return nil
		},
	}
}

func NewRootCommand(opts *options) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "relayer",
		Version:       "DEV",
		Short:         "",
		Long:          ``,
		SilenceErrors: false,
		SilenceUsage:  true,
	}

	rootCmd.PersistentFlags().StringVarP(&opts.ConfigFilePath, "relayer-config", "c", "./relayer.json", "config file")

	return rootCmd
}
