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

	ghostConfig "github.com/makerdao/gofer/pkg/ghost/config"
	goferConfig "github.com/makerdao/gofer/pkg/gofer/config"
)

type options struct {
	LogVerbosity        string
	GhostConfigFilePath string
	GoferConfigFilePath string
	GhostConfig         ghostConfig.Config
	GoferConfig         goferConfig.Config
}

func NewRootCommand(opts *options) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "ghost",
		Version:       "DEV",
		Short:         "",
		Long:          ``,
		SilenceErrors: false,
		SilenceUsage:  true,
	}

	rootCmd.PersistentFlags().StringVarP(
		&opts.LogVerbosity,
		"log.verbosity", "v",
		"info",
		"verbosity level",
	)
	rootCmd.PersistentFlags().StringVarP(
		&opts.GhostConfigFilePath,
		"config", "c",
		"./ghost.json",
		"ghost config file",
	)
	rootCmd.PersistentFlags().StringVar(
		&opts.GoferConfigFilePath,
		"config.gofer",
		"./gofer.json",
		"gofer config file",
	)

	return rootCmd
}
