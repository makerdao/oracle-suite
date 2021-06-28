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
	configCobra "github.com/makerdao/oracle-suite/pkg/spire/config/cobra"
	"github.com/spf13/cobra"
)

func NewRootCommand(opts *options) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "spamman",
		Version: opts.Version,
		Short:   "Tool for testing system by spamming invalid/valid messages",
		Long: `
Spamman is a CLI interface for the Spamman Test Utility.

It is a tool that allows easy generating/sending spam messages.`,
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	rootCmd.PersistentFlags().StringVarP(
		&opts.LogVerbosity,
		"log.verbosity", "v",
		"info",
		"verbosity level",
	)
	rootCmd.PersistentFlags().Var(
		&opts.LogFormat,
		"log.format",
		"log format",
	)
	rootCmd.PersistentFlags().StringVarP(
		&opts.ConfigPath,
		"config",
		"c",
		"./spire.json",
		"config file (similar to spire)",
	)
	rootCmd.PersistentFlags().IntVarP(
		&opts.MessageRate,
		"msg.rate",
		"r",
		20,
		"message rate (msg per min)",
	)
	rootCmd.PersistentFlags().BoolVar(
		&opts.ValidMessages,
		"gen.valid.msg",
		false,
		"valid price messages will be generated",
	)
	rootCmd.PersistentFlags().BoolVar(
		&opts.InvalidSignature,
		"gen.invalid.signature",
		false,
		"messages with invalid signature will be generated",
	)

	configCobra.RegisterFlags(&opts.Config, rootCmd.PersistentFlags())

	rootCmd.AddCommand(
		NewRunCmd(opts),
	)

	return rootCmd
}
