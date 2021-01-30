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
