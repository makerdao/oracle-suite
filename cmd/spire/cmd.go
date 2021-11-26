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
	"github.com/spf13/cobra"

	logrusFlag "github.com/makerdao/oracle-suite/pkg/log/logrus/flag"
)

type options struct {
	LogVerbosity   logrusFlag.Verbosity
	LogFormat      logrusFlag.Format
	ConfigFilePath string
	Config         Config
	Version        string
	Transport      string
}

func NewRootCommand(opts *options) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "spire",
		Version:       opts.Version,
		Short:         "",
		Long:          "",
		SilenceErrors: false,
		SilenceUsage:  true,
	}

	rootCmd.PersistentFlags().VarP(
		&opts.LogVerbosity,
		"log.verbosity",
		"v",
		"verbosity level",
	)
	rootCmd.PersistentFlags().Var(
		&opts.LogFormat,
		"log.format",
		"log format",
	)
	rootCmd.PersistentFlags().StringVarP(
		&opts.ConfigFilePath,
		"config",
		"c",
		"./config.json",
		"spire config file",
	)
	rootCmd.PersistentFlags().StringVar(
		&opts.Transport,
		"transport",
		"libp2p",
		"`libp2p|ssb` transport spire connects to",
	)

	rootCmd.AddCommand(
		NewAgentCmd(opts),
		NewPullCmd(opts),
		NewPushCmd(opts),
	)

	return rootCmd
}
